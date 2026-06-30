package seeders

import (
	"fmt"
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
)

func ResetDatabases(app *global.App) error {
	if app == nil || app.DB == nil {
		return fmt.Errorf("app or postgres connection is nil")
	}

	if err := app.DB.Exec("DROP SCHEMA IF EXISTS public CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop public schema: %w", err)
	}

	if err := app.DB.Exec("CREATE SCHEMA public").Error; err != nil {
		return fmt.Errorf("failed to create public schema: %w", err)
	}

	if err := app.DB.Exec("GRANT ALL ON SCHEMA public TO public").Error; err != nil {
		return fmt.Errorf("failed to grant schema permissions: %w", err)
	}

	// mongoClient := global.GetMongo()
	// mongoName := global.GetMongoDBName()
	// if err := mongoClient.Database(mongoName).Drop(context.Background()); err != nil {
	// 	return fmt.Errorf("failed to drop mongo database %q: %w", mongoName, err)
	// }

	log.Printf("Reset complete for Postgres schema")
	return nil
}
