package cmd

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/api"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	ledgergrpc "github.com/Cork-Holdings/gp_payment_orchestration/internal/grpc"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/approvals"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/ledger"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/mq"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/tasks"
	"github.com/Cork-Holdings/gp_payment_orchestration/proto/ledgerpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve app",
	Long:  `command is used to serve the app`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting api gateway")

		app := global.New()

		// Register and migrate ledger and approvals models
		app.Register(&ledger.Account{}, &approvals.ApprovalRequest{})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup

		// Start gRPC Server
		lis, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("Failed to listen on port 50052: %v", err)
		}
		grpcServer := grpc.NewServer()
		ledgerpb.RegisterLedgerServiceServer(grpcServer, ledgergrpc.NewLedgerServer(app))

		go func() {
			log.Println("Starting gRPC server on port 50052...")
			if err := grpcServer.Serve(lis); err != nil {
				log.Printf("gRPC server stopped: %v", err)
				cancel()
			}
		}()

		// API Server
		go func() {
			if err := api.Server(app); err != nil {
				log.Printf("server stopped: %v", err)
				cancel()
			}
		}()

		wg.Add(1)

		// Background Tasks
		go func() {
			jobQueue, err := tasks.Run(app)
			if err != nil {
				log.Printf("task runner stopped: %v", err)
				cancel()
				return
			}

			<-ctx.Done()

			log.Println("Shutting down job queue...")
			jobQueue.Shutdown()
		}()
		wg.Add(1)

		// MQ Consumer
		go func() {
			if err := app.MQ.Consume(
				app,
				os.Getenv("QUEUE_NAME"),
				mq.Reciever,
			); err != nil {
				log.Printf("consumer stopped: %v", err)
				cancel()
			}
		}()

		// Listen for Ctrl+C / kill
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		select {
		case sig := <-sigChan:
			log.Printf("Received signal %s. Shutting down...", sig)
			cancel()

		case <-ctx.Done():
			log.Println("Application stopping...")
		}

		log.Println("Gracefully stopping gRPC server...")
		grpcServer.GracefulStop()
		app.Close()

		log.Println("Shutdown complete")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
