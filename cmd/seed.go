package cmd

import (
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/database/seeders"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with initial data",
	Run: func(cmd *cobra.Command, args []string) {
		app := global.New()
		RegisterAppModels(app)
		if err := seeders.Seed(app.DB); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
}
