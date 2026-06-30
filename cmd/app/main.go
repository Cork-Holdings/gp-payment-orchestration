package main

import (
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/cmd"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Failed to load env: %v", err.Error())
	}

	app := global.New()
	cmd.RegisterAppModels(app)

	// mocks.StartFakeTransactionService(app)
	// mocks.StartFakeMerchantService(app)
	// mocks.StartMerchantCollectionBalanceService(app)
	// mocks.StartMerchantDisbursementBalanceService(app)

	// mocks.StartFakeCollectionsService("test-collection-ref-01", app)
	// mocks.StartFakeDisbursementsService("test-disbursement-ref-01", app)

	cmd.Execute()
}
