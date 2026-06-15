package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/api"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/mq"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/tasks"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve app",
	Long:  `command is used to serve the app`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting api gateway")

		app := global.New()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup

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

		app.Close()

		log.Println("Shutdown complete")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
