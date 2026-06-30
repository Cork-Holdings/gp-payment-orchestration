package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/api"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/database/seeders"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/mq"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/tasks"
	"github.com/spf13/cobra"
)

func envBool(keys ...string) bool {
	for _, key := range keys {
		val := strings.TrimSpace(os.Getenv(key))
		if val == "" {
			continue
		}

		parsed, err := strconv.ParseBool(strings.ToLower(val))
		if err == nil {
			return parsed
		}

		switch strings.ToLower(val) {
		case "yes", "y", "on":
			return true
		case "no", "n", "off":
			return false
		}
	}

	return false
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve app",
	Long:  `command is used to serve the app`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting api gateway")

		app := global.New()
		RegisterAppModels(app)

		shouldReset := envBool("RESET", "reset")
		shouldSeed := shouldReset || envBool("SEED", "seed")

		if shouldReset {
			log.Println("RESET=true detected. Resetting Postgres schema and Mongo database...")
			if err := seeders.ResetDatabases(app); err != nil {
				log.Fatalf("failed to reset data stores: %v", err)
			}

			// Re-apply schema after reset.
			app.Models = nil
			RegisterAppModels(app)
		}

		if shouldSeed {
			log.Println("Seeding enabled via environment variable")
			if err := seeders.Seed(app.DB); err != nil {
				log.Fatalf("failed to seed database: %v", err)
			}
		}

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
