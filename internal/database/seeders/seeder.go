package seeders

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapikeys"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantfeeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantips"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantpaymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
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
	db.Exec("DELETE FROM merchant_api_keys")
	db.Exec("DELETE FROM merchant_ips")
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
	db.Exec("DELETE FROM providers")

	// 1. Providers
	provAirtel := providers.Provider{ID: uuid.New(), Name: "Airtel", Code: "AIRTEL", Status: "active"}
	provMTN := providers.Provider{ID: uuid.New(), Name: "MTN", Code: "MTN", Status: "active"}
	provZamtel := providers.Provider{ID: uuid.New(), Name: "Zamtel", Code: "ZAMTEL", Status: "active"}
	provZedmobile := providers.Provider{ID: uuid.New(), Name: "ZEDMOBILE", Code: "ZEDMOBILE", Status: "active"}

	mobileProviders := []providers.Provider{provAirtel, provMTN, provZamtel, provZedmobile}
	if err := db.Create(&mobileProviders).Error; err != nil {
		return err
	}

	// Zambian Banks
	banks := []string{
		"FNB Zambia",
		"Zanaco",
		"Standard Chartered Bank Zambia",
		"Absa Bank Zambia",
		"Access Bank Zambia",
		"Stanbic Bank Zambia",
		"Investrust Bank",
		"Ecobank Zambia",
		"United Bank for Africa (UBA)",
		"Bank of China Zambia",
	}

	for _, bankName := range banks {
		bankCode := strings.ToUpper(strings.ReplaceAll(bankName, " ", "_"))
		if err := db.Create(&providers.Provider{
			ID:     uuid.New(),
			Name:   bankName,
			Code:   bankCode,
			Status: "active",
		}).Error; err != nil {
			return err
		}
	}

	// 2. Subscriptions
	subMM := subscriptions.Subscription{
		ID:          uuid.New(),
		Name:        "Mobile Money",
		Status:      "active",
		Description: "Give merchant access to Mobile Money transactions, collection and disbursement",
	}
	if err := db.Create(&subMM).Error; err != nil {
		return err
	}

	// 3. Transaction Types
	ttCollection := transactiontypes.TransactionType{
		ID:        uuid.New(),
		Name:      "MNO Collection",
		Code:      "MNO_COLLECTION",
		MaxAmount: "1000000",
		MinAmount: "0.01",
		Status:    "active",
	}
	ttDisbursement := transactiontypes.TransactionType{
		ID:        uuid.New(),
		Name:      "MNO Disbursement",
		Code:      "MNO_DISBURSEMENT",
		MaxAmount: "1000000",
		MinAmount: "0.01",
		Status:    "active",
	}
	if err := db.Create(&[]transactiontypes.TransactionType{ttCollection, ttDisbursement}).Error; err != nil {
		return err
	}

	// 4. Payment Channels
	channels := []paymentchannels.PaymentChannel{
		{ID: uuid.New(), Name: "Airtel (Collection)", Code: "airtel", ProviderID: provAirtel.ID, SubscriptionID: subMM.ID, TransactionTypeID: ttCollection.ID, FeeType: "band", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "Airtel (Disbursement)", Code: "airtel_disburse", ProviderID: provAirtel.ID, SubscriptionID: subMM.ID, TransactionTypeID: ttDisbursement.ID, FeeType: "percentage", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "MTN (Collection)", Code: "mtn", ProviderID: provMTN.ID, SubscriptionID: subMM.ID, TransactionTypeID: ttCollection.ID, FeeType: "band", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "MTN (Disbursement)", Code: "mtn_disburse", ProviderID: provMTN.ID, SubscriptionID: subMM.ID, TransactionTypeID: ttDisbursement.ID, FeeType: "percentage", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "Zamtel (Collection)", Code: "zamtel", ProviderID: provZamtel.ID, SubscriptionID: subMM.ID, TransactionTypeID: ttCollection.ID, FeeType: "band", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "Zamtel (Disbursement)", Code: "zamtel_disburse", ProviderID: provZamtel.ID, SubscriptionID: subMM.ID, TransactionTypeID: ttDisbursement.ID, FeeType: "percentage", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "ZEDMOBILE (Collection)", Code: "zedmobile", ProviderID: provZedmobile.ID, SubscriptionID: subMM.ID, TransactionTypeID: ttCollection.ID, FeeType: "band", ProviderFee: "0.00", Status: "active"},
		{ID: uuid.New(), Name: "ZEDMOBILE (Disbursement)", Code: "zedmobile_disburse", ProviderID: provZedmobile.ID, SubscriptionID: subMM.ID, TransactionTypeID: ttDisbursement.ID, FeeType: "percentage", ProviderFee: "0.00", Status: "active"},
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

	// 5. Channel Fee Bands (Provider Fees - MNO costs)
	// Realistic bands based on industry standards for Airtel, MTN, Zamtel, ZEDMOBILE
	var bands []paymentchannels.ChannelFeeBands

	collectionChannels := []string{"MTN (Collection)", "Airtel (Collection)", "Zamtel (Collection)", "ZEDMOBILE (Collection)"}

	feeBands := [][]float64{
		{1, 150, 0.50},
		{151, 300, 1.00},
		{301, 500, 1.00},
		{501, 1000, 1.50},
		{1001, 3000, 2.80},
		{3001, 5000, 4.00},
		{5001, 10000, 5.50},
	}

	for _, cName := range collectionChannels {
		if cID, ok := channelMap[cName]; ok {
			for _, b := range feeBands {
				bands = append(bands, paymentchannels.ChannelFeeBands{
					ID:               uuid.New(),
					Name:             fmt.Sprintf("%s %.0f-%.0f", cName, b[0], b[1]),
					PaymentChannelID: cID,
					MinAmount:        b[0],
					MaxAmount:        b[1],
					ChargeAmount:     b[2],
					ChargeType:       "fixed",
					Status:           "active",
				})
			}
		}
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
		{"26078", []string{"ZEDMOBILE (Collection)", "ZEDMOBILE (Disbursement)"}},
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
				AssignedBy:       uuid.New(),
				ApprovedBy:       uuid.New(),
			})

			// Create a fee profile for each merchant for each channel
			// Use different profile types for different merchants to test variety
			var fp feeprofiles.FeeProfile
			switch mID {
			case merchantID2:
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
					// Tiered structure: 0-100 K1 fixed, 100-500 1.5%, 500+ 1%
					db.Create(&feeprofiles.ProfileFeeBands{ID: uuid.New(), FeeProfileID: fp.ID, MinAmount: 0.01, MaxAmount: 100, ChargeAmount: 1, ChargeType: "fixed", Status: "active"})
					db.Create(&feeprofiles.ProfileFeeBands{ID: uuid.New(), FeeProfileID: fp.ID, MinAmount: 100.01, MaxAmount: 500, ChargeAmount: 1.5, ChargeType: "percentage", Status: "active"})
					db.Create(&feeprofiles.ProfileFeeBands{ID: uuid.New(), FeeProfileID: fp.ID, MinAmount: 500.01, MaxAmount: 0, ChargeAmount: 1.0, ChargeType: "percentage", Status: "active"})
				} else {
					// Disbursement: 1.2% with K1 minimum
					fp = feeprofiles.FeeProfile{
						ID:                uuid.New(),
						Name:              mID.String()[:8] + " " + c.Name + " Std",
						Code:              "STD_" + c.Code,
						PaymentChannelID:  c.ID,
						TransactionTypeID: c.TransactionTypeID,
						Status:            "active",
						ChargeType:        "percentage",
						ChargeAmount:      1.2,
						MinimumFee:        1,
						ApprovalStatus:    "approved",
						CalculationMode:   "standard",
					}
					db.Create(&fp)
				}
			case merchantID3:
				// Merchant 3: Percentage for all channels (realistic rates)
				chargeAmount := 2.0 // 2% for collection
				minimumFee := 1.0
				if c.TransactionTypeID == ttDisbursement.ID {
					chargeAmount = 1.1 // 1.1% for disbursement
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
					MinimumFee:        minimumFee,
					ApprovalStatus:    "approved",
					CalculationMode:   "standard",
				}
				db.Create(&fp)
			default:
				// Other merchants (1 & 4): Realistic percentage fees
				chargeAmount := 1.8 // 1.8% for collection
				minimumFee := 1.0
				if c.TransactionTypeID == ttDisbursement.ID {
					chargeAmount = 1.0 // 1% for disbursement
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
					MinimumFee:        minimumFee,
					ApprovalStatus:    "approved",
					CalculationMode:   "standard",
				}
				db.Create(&fp)
			}
			db.Create(&merchantfeeprofiles.MerchantFeeProfile{ID: uuid.New(), MerchantID: mID, FeeProfileID: fp.ID, Status: "active"})
		}
		db.Create(&subscriptions.MerchantSubscription{ID: uuid.New(), MerchantID: mID, SubscriptionID: subMM.ID, Status: "active"})

		// 9. Seed API Keys and provision merchant accounts in transactions service.
		if _, err := merchantapikeys.CreateMerchantKeys(mID.String()); err != nil {
			return err
		}

		db.Create(&merchantips.MerchantIP{
			ID:          uuid.New(),
			MerchantID:  mID,
			IPAddress:   "127.0.0.1",
			Status:      "approved",
			SubmittedBy: mID,
		})
		// Also add IPv6 localhost for local testing
		//Testing webhook
		db.Create(&merchantips.MerchantIP{
			ID:          uuid.New(),
			MerchantID:  mID,
			IPAddress:   "::1",
			Status:      "approved",
			SubmittedBy: mID,
		})
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
