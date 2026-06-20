package seeders

import (
	"context"
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantfeeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantpaymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentservices"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/prefixes"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/providers"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/subscriptions"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/transactiontypes"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	log.Println("Seeding database...")

	// Clear cache
	cache := global.GetCache()
	if cache != nil {
		cache.FlushAll(context.Background())
		log.Println("Cache cleared.")
	}

	// Clear existing data
	db.Exec("DELETE FROM merchant_subscriptions")
	db.Exec("DELETE FROM subscriptions")
	db.Exec("DELETE FROM prefix_payment_channels")
	db.Exec("DELETE FROM prefixes")
	db.Exec("DELETE FROM profile_fee_bands")
	db.Exec("DELETE FROM merchant_fee_profiles")
	db.Exec("DELETE FROM fee_profiles")
	db.Exec("DELETE FROM channel_fee_bands")
	db.Exec("DELETE FROM merchant_payment_channels")
	db.Exec("DELETE FROM payment_channels")
	db.Exec("DELETE FROM sub_transaction_types")
	db.Exec("DELETE FROM transaction_types")
	db.Exec("DELETE FROM payment_services")
	db.Exec("DELETE FROM providers")

	// 1. Providers
	provAirtel := providers.Provider{ID: uuid.New(), Name: "Airtel", Code: "AIRTEL", Status: "active"}
	provMTN := providers.Provider{ID: uuid.New(), Name: "MTN", Code: "MTN", Status: "active"}
	provZamtel := providers.Provider{ID: uuid.New(), Name: "Zamtel", Code: "ZAMTEL", Status: "active"}
	if err := db.Create(&[]providers.Provider{provAirtel, provMTN, provZamtel}).Error; err != nil {
		return err
	}

	// 2. Payment Services
	psMM := paymentservices.PaymentService{
		ID:     uuid.New(),
		Name:   "Mobile Money",
		Status: "active",
		Logo:   "https://example.com/mobile_money.png",
	}
	if err := db.Create(&psMM).Error; err != nil {
		return err
	}

	// 3. Transaction Types
	ttCollection := transactiontypes.TransactionType{
		ID:        uuid.New(),
		Name:      "Collection",
		Code:      "COLLECTION",
		MaxAmount: "1000000",
		MinAmount: "0.01",
		Status:    "active",
	}
	ttDisbursement := transactiontypes.TransactionType{
		ID:        uuid.New(),
		Name:      "Disbursement",
		Code:      "DISBURSEMENT",
		MaxAmount: "1000000",
		MinAmount: "0.01",
		Status:    "active",
	}
	if err := db.Create(&[]transactiontypes.TransactionType{ttCollection, ttDisbursement}).Error; err != nil {
		return err
	}

	// 4. Payment Channels
	channels := []paymentchannels.PaymentChannel{
		{ID: uuid.New(), Name: "Airtel (Collection)", Code: "airtel", ProviderID: provAirtel.ID, PaymentServiceID: psMM.ID, TransactionTypeID: ttCollection.ID, FeeType: "band", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "Airtel (Disbursement)", Code: "airtel_disburse", ProviderID: provAirtel.ID, PaymentServiceID: psMM.ID, TransactionTypeID: ttDisbursement.ID, FeeType: "percentage", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "MTN (Collection)", Code: "mtn", ProviderID: provMTN.ID, PaymentServiceID: psMM.ID, TransactionTypeID: ttCollection.ID, FeeType: "band", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "MTN (Disbursement)", Code: "mtn_disburse", ProviderID: provMTN.ID, PaymentServiceID: psMM.ID, TransactionTypeID: ttDisbursement.ID, FeeType: "percentage", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "Zamtel (Collection)", Code: "zamtel", ProviderID: provZamtel.ID, PaymentServiceID: psMM.ID, TransactionTypeID: ttCollection.ID, FeeType: "band", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "Zamtel (Disbursement)", Code: "zamtel_disburse", ProviderID: provZamtel.ID, PaymentServiceID: psMM.ID, TransactionTypeID: ttDisbursement.ID, FeeType: "percentage", ProviderFee: "0.00", Status: "active"},
	}
	for i := range channels {
		if err := db.Create(&channels[i]).Error; err != nil {
			return err
		}
	}

	// Map channels for easy access
	channelMap := make(map[string]uuid.UUID)
	for _, c := range channels {
		channelMap[c.Name] = c.ID
	}

	// 5. Channel Fee Bands (Provider Fees)
	bands := []paymentchannels.ChannelFeeBands{
		// MTN (Collection)
		{ID: uuid.New(), Name: "MTN Collection Band 1", PaymentChannelID: channelMap["MTN (Collection)"], MinAmount: 3000.01, MaxAmount: 5000.00, ChargeAmount: 3.00, ChargeType: "fixed", Status: "active"},
		{ID: uuid.New(), Name: "MTN Collection Band 2", PaymentChannelID: channelMap["MTN (Collection)"], MinAmount: 1000.01, MaxAmount: 3000.00, ChargeAmount: 2.20, ChargeType: "fixed", Status: "active"},
		{ID: uuid.New(), Name: "MTN Collection Band 3", PaymentChannelID: channelMap["MTN (Collection)"], MinAmount: 300.01, MaxAmount: 500.00, ChargeAmount: 0.80, ChargeType: "fixed", Status: "active"},
		{ID: uuid.New(), Name: "MTN Collection Band 4", PaymentChannelID: channelMap["MTN (Collection)"], MinAmount: 150.01, MaxAmount: 300.00, ChargeAmount: 0.90, ChargeType: "fixed", Status: "active"},
		{ID: uuid.New(), Name: "MTN Collection Band 5", PaymentChannelID: channelMap["MTN (Collection)"], MinAmount: 0.01, MaxAmount: 150.00, ChargeAmount: 0.42, ChargeType: "fixed", Status: "active"},

		// Airtel (Collection)
		{ID: uuid.New(), Name: "Airtel Collection Band 1", PaymentChannelID: channelMap["Airtel (Collection)"], MinAmount: 150.01, MaxAmount: 500.00, ChargeAmount: 1.00, ChargeType: "fixed", Status: "active"},
		{ID: uuid.New(), Name: "Airtel Collection Band 2", PaymentChannelID: channelMap["Airtel (Collection)"], MinAmount: 1000.01, MaxAmount: 3000.00, ChargeAmount: 2.80, ChargeType: "fixed", Status: "active"},
		{ID: uuid.New(), Name: "Airtel Collection Band 3", PaymentChannelID: channelMap["Airtel (Collection)"], MinAmount: 500.01, MaxAmount: 1000.00, ChargeAmount: 1.50, ChargeType: "fixed", Status: "active"},

		// Zamtel (Collection)
		{ID: uuid.New(), Name: "Zamtel Collection Band 1", PaymentChannelID: channelMap["Zamtel (Collection)"], MinAmount: 150.01, MaxAmount: 300.00, ChargeAmount: 0.89, ChargeType: "fixed", Status: "active"},
		{ID: uuid.New(), Name: "Zamtel Collection Band 2", PaymentChannelID: channelMap["Zamtel (Collection)"], MinAmount: 0.00, MaxAmount: 4.99, ChargeAmount: 0.00, ChargeType: "fixed", Status: "active"},
	}
	if err := db.Create(&bands).Error; err != nil {
		return err
	}

	// 6. Prefixes
	prefixData := []struct {
		Prefix   string
		Channels []string
	}{
		{"26075", []string{"Zamtel (Collection)", "Zamtel (Disbursement)"}},
		{"26057", []string{"Airtel (Collection)", "Airtel (Disbursement)"}},
		{"26097", []string{"Airtel (Collection)", "Airtel (Disbursement)"}},
		{"26077", []string{"Airtel (Collection)", "Airtel (Disbursement)"}},
		{"26096", []string{"MTN (Collection)", "MTN (Disbursement)"}},
		{"26076", []string{"MTN (Collection)", "MTN (Disbursement)"}},
		{"26095", []string{"Zamtel (Collection)", "Zamtel (Disbursement)"}},
	}

	for _, p := range prefixData {
		pre := prefixes.Prefix{ID: uuid.New(), Prefix: p.Prefix}
		if err := db.Create(&pre).Error; err != nil {
			return err
		}
		for _, cName := range p.Channels {
			if cID, ok := channelMap[cName]; ok {
				db.Create(&prefixes.PrefixPaymentChannel{
					ID:               uuid.New(),
					PrefixID:         pre.ID,
					PaymentChannelID: cID,
				})
			}
		}
	}

	// 7. Subscriptions
	subMM := subscriptions.Subscription{
		ID:          uuid.New(),
		Name:        "Mobile Money",
		Status:      "active",
		Description: "Give merchant access to Mobile Money transactions, collection and disbursement",
	}
	if err := db.Create(&subMM).Error; err != nil {
		return err
	}

	// 8. Create Merchants and link them to channels and profiles
	merchantID1 := uuid.MustParse("3d15c820-485c-4ceb-b040-4de1e7d4bd56")
	merchantID2 := uuid.MustParse("5859ed51-5aad-4920-8980-8447563d2b04")
	merchantID3 := uuid.MustParse("6ef69e57-f112-4cde-a3da-0e6b1cb4521f")
	merchantID4 := uuid.New()
	allMerchants := []uuid.UUID{merchantID1, merchantID2, merchantID3, merchantID4}

	for _, mID := range allMerchants {
		// Authorize all merchants for all channels
		for _, c := range channels {
			db.Create(&merchantpaymentchannels.MerchantPaymentChannel{
				ID:               uuid.New(),
				MerchantID:       mID,
				PaymentChannelID: c.ID,
				Status:           "active",
				ApprovalStatus:   "approved",
			})

			// Create a fee profile for each merchant for each channel
			// Use different profile types for different merchants to test variety
			var fp feeprofiles.FeeProfile
			if mID == merchantID2 {
				// Merchant 2: Tiered/Band for all collection channels
				if c.TransactionTypeID == ttCollection.ID {
					fp = feeprofiles.FeeProfile{
						ID:                uuid.New(),
						Name:              mID.String()[:8] + " " + c.Name + " Tiered",
						Code:              "TIERED_" + c.Code,
						PaymentChannelID:  c.ID,
						TransactionTypeID: c.TransactionTypeID,
						Status:            "active",
						ChargeType:        "band",
						ApprovalStatus:    "approved",
						CalculationMode:   "standard",
					}
					db.Create(&fp)
					db.Create(&feeprofiles.ProfileFeeBands{ID: uuid.New(), FeeProfileID: fp.ID, MinAmount: 0, MaxAmount: 50, ChargeAmount: 2, ChargeType: "fixed", Status: "active"})
					db.Create(&feeprofiles.ProfileFeeBands{ID: uuid.New(), FeeProfileID: fp.ID, MinAmount: 50.01, MaxAmount: 1000000, ChargeAmount: 1, ChargeType: "percentage", Status: "active"})
				} else {
					// Disbursement: Standard percentage
					fp = feeprofiles.FeeProfile{
						ID:                uuid.New(),
						Name:              mID.String()[:8] + " " + c.Name + " Std",
						Code:              "STD_" + c.Code,
						PaymentChannelID:  c.ID,
						TransactionTypeID: c.TransactionTypeID,
						Status:            "active",
						ChargeAmount:      2,
						MinimumFee:        5,
						ApprovalStatus:    "approved",
						CalculationMode:   "standard",
					}
					db.Create(&fp)
				}
			} else if mID == merchantID3 {
				// Merchant 3: Percentage for all collection channels
				chargeAmount := 3.5 // 3.5%
				if c.TransactionTypeID == ttDisbursement.ID {
					chargeAmount = 2.0
				}
				fp = feeprofiles.FeeProfile{
					ID:                uuid.New(),
					Name:              mID.String()[:8] + " " + c.Name + " Percentage",
					Code:              "PERCENT_" + c.Code,
					PaymentChannelID:  c.ID,
					TransactionTypeID: c.TransactionTypeID,
					Status:            "active",
					ChargeType:        "percentage",
					ChargeAmount:      chargeAmount,
					MinimumFee:        5,
					ApprovalStatus:    "approved",
					CalculationMode:   "standard",
				}
				db.Create(&fp)
			} else {
				// Other merchants: Fixed
				chargeAmount := 20.0
				if c.TransactionTypeID == ttDisbursement.ID {
					chargeAmount = 2.0
				}
				fp = feeprofiles.FeeProfile{
					ID:                uuid.New(),
					Name:              mID.String()[:8] + " " + c.Name + " Fixed/Std",
					Code:              "FIXED_" + c.Code,
					PaymentChannelID:  c.ID,
					TransactionTypeID: c.TransactionTypeID,
					Status:            "active",
					ChargeAmount:      chargeAmount,
					MinimumFee:        5,
					ApprovalStatus:    "approved",
					CalculationMode:   "fixed",
				}
				if c.TransactionTypeID == ttDisbursement.ID {
					fp.CalculationMode = "standard"
					fp.ChargeType = "percentage"
				}
				db.Create(&fp)
			}
			db.Create(&merchantfeeprofiles.MerchantFeeProfile{ID: uuid.New(), MerchantID: mID, FeeProfileID: fp.ID, Status: "active"})
		}
		db.Create(&subscriptions.MerchantSubscription{ID: uuid.New(), MerchantID: mID, SubscriptionID: subMM.ID, Status: "active"})
	}

	log.Printf("\n--- SEED DATA FOR SIMULATION ---\n")
	log.Printf("Merchant 1 (Fixed): %s\n", merchantID1)
	log.Printf("Merchant 2 (Tiered): %s\n", merchantID2)
	log.Printf("Merchant 3 (Percentage): %s\n", merchantID3)
	log.Printf("Merchant 4 (Fixed): %s\n", merchantID4)
	log.Printf("--------------------------------\n")

	log.Println("Database seeded successfully!")
	return nil
}
