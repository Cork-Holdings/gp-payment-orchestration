package main

import (
	"fmt"
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feecalculator"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantfeeprofiles"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment and initialize App
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	db := global.GetDB()

	fmt.Println("=== COMPREHENSIVE FEE CALCULATION SIMULATION ===")

	// 2. Find test merchants from DB
	var mfps []merchantfeeprofiles.MerchantFeeProfile
	if err := db.Preload("FeeProfile").Find(&mfps).Error; err != nil {
		log.Fatalf("Failed to fetch merchant fee profiles: %v", err)
	}

	if len(mfps) == 0 {
		fmt.Println("No data found. Please run 'go run cmd/app/main.go seed' first.")
		return
	}

	// 3. Define prefixes for different providers
	prefixes := map[string]string{
		"Airtel": "26077",
		"MTN":    "26096",
		"Zamtel": "26075",
	}

	// 4. Define test amounts to hit different bands
	// MTN bands: 0-150, 150-300, 300-500, 1000-3000, 3000-5000
	// Airtel bands: 150-500, 500-1000, 1000-3000
	// Zamtel bands: 0-4.99, 150-300
	testAmounts := []float64{2.50, 
		30.00, 
		100.00, 
		200.00, 
		400.00, 
		750.00, 
		1500.00, 
		4000.00,
		10000.00,
	}

	// 5. Run simulations for all merchants
	for _, mfp := range mfps {
		merchantID := mfp.MerchantID.String()
		profileName := mfp.FeeProfile.Name
		profileCode := mfp.FeeProfile.Code

		fmt.Printf("\n>>> Testing Merchant: %s (Profile: %s, Code: %s) <<<\n", merchantID, profileName, profileCode)

		// Determine which providers/prefixes to test based on merchant profile or just test all
		for provider, prefix := range prefixes {
			phone := prefix + "123456"

			// Test both Collection and Disbursement if applicable
			txnTypes := []feecalculator.TransactionTypeCode{
				feecalculator.TransactionTypeCollection,
				feecalculator.TransactionTypeDisbursement,
			}

			for _, txnType := range txnTypes {
				fmt.Printf("\n--- Provider: %s, Type: %s ---\n", provider, txnType)

				for _, amount := range testAmounts {
					fmt.Printf("Input: Amount=%.2f\n", amount)

					result, err := feecalculator.CalculateFees(merchantID, phone, amount, txnType)
					if err != nil {
						fmt.Printf("  Error: %v\n", err)
						continue
					}

					if result.Error != "" {
						fmt.Printf("  Result Error: %s\n", result.Error)
						continue
					}

					// Print key results concisely
					fmt.Printf("  Fee: %.2f (Profile: %s, Provider Band: %s)\n", 
						result.TransactionFeeAmount, result.ProfileFeeBandRange, result.ProviderFeeBandRange)
					
					// If it's a specific case we want to see full JSON for, we can do it here
					// For now, let's just show the summary to keep output readable
				}
			}
		}
	}

	fmt.Println("\n=== SIMULATION COMPLETE ===")
}
