package seeders

import (
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantfeeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentservices"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/prefixes"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/subscriptions"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/transactiontypes"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	log.Println("Seeding database...")

	// 1. Payment Services
	ps1 := paymentservices.PaymentService{
		ID:     uuid.New(),
		Name:   "Card",
		Status: "active",
		Logo:   "https://example.com/mpesa.png",
	}
	ps2 := paymentservices.PaymentService{
		ID:     uuid.New(),
		Name:   "Mobile Money",
		Status: "active",
		Logo:   "https://example.com/airtel.png",
	}
	ps3 := paymentservices.PaymentService{
		ID:     uuid.New(),
		Name:   "NFS",
		Status: "active",
		Logo:   "https://example.com/airtel.png",
	}
	if err := db.Create(&[]paymentservices.PaymentService{ps1, ps2, ps3}).Error; err != nil {
		return err
	}

	// 2. Transaction Types
	tt1 := transactiontypes.TransactionType{
		ID:        uuid.New(),
		Name:      "Collection",
		Code:      "COLLECTION",
		MaxAmount: "1000000",
		MinAmount: "10",
		Status:    "active",
	}
	tt2 := transactiontypes.TransactionType{
		ID:        uuid.New(),
		Name:      "Disbursement",
		Code:      "DISBURSEMENT",
		MaxAmount: "1000000",
		MinAmount: "10",
		Status:    "active",
	}
	tt3 := transactiontypes.TransactionType{
		ID:        uuid.New(),
		Name:      "E-Money",
		Code:      "E_MONEY",
		MaxAmount: "0",
		MinAmount: "0",
		Status:    "active",
	}
	if err := db.Create(&[]transactiontypes.TransactionType{tt1, tt2, tt3}).Error; err != nil {
		return err
	}

	// 3. Sub Transaction Types
	stt1 := transactiontypes.SubTransactionType{
		ID:                uuid.New(),
		Name:              "Cash In",
		TransactionTypeID: tt3.ID,
		Code:              "CASH_IN",
		Status:            "active",
		MaxAmount:         "1000000",
		MinAmount:         "10",
	}
	if err := db.Create(&stt1).Error; err != nil {
		return err
	}

	// 4. Payment Channels
	pc1 := paymentchannels.PaymentChannel{
		ID:                uuid.New(),
		Name:              "AirTel Money",
		Code:              "AIRTEL_MONEY",
		Status:            "active",
		PaymentServiceID:  ps1.ID,
		TransactionTypeID: tt1.ID,
		FeeType:           "percentage",
		ProviderFee:       "1.5",
	}
	if err := db.Create(&pc1).Error; err != nil {
		return err
	}

	// 5. Channel Fee Bands
	cfb1 := paymentchannels.ChannelFeeBands{
		ID:               uuid.New(),
		Name:             "Standard Collection Band",
		PaymentChannelID: pc1.ID,
		MinAmount:        0,
		MaxAmount:        1000000,
		ChargeAmount:     10,
		Status:           "active",
	}
	if err := db.Create(&cfb1).Error; err != nil {
		return err
	}

	// 6. Fee Profiles
	fp1 := feeprofiles.FeeProfile{
		ID:                uuid.New(),
		Name:              "Standard Collection Profile",
		Code:              "STD_COLLECTION",
		PaymentChannelID:  pc1.ID,
		TransactionTypeID: tt1.ID,
		Status:            "active",
		ChargeAmount:      20,
		ApprovalStatus:    "approved",
		CalculationMode:   "fixed",
	}
	if err := db.Create(&fp1).Error; err != nil {
		return err
	}

	// 7. Profile Fee Bands
	pfb1 := feeprofiles.ProfileFeeBands{
		ID:           uuid.New(),
		FeeProfileID: fp1.ID,
		MinAmount:    0,
		MaxAmount:    1000000,
		ChargeAmount: 5,
		ChargeType:   "fixed",
		Status:       "active",
	}
	if err := db.Create(&pfb1).Error; err != nil {
		return err
	}

	// 8. Prefixes
	pre1 := prefixes.Prefix{
		ID:     uuid.New(),
		Prefix: "26077",
	}
	if err := db.Create(&pre1).Error; err != nil {
		return err
	}

	// 9. Prefix Payment Channels
	ppc1 := prefixes.PrefixPaymentChannel{
		ID:               uuid.New(),
		PrefixID:         pre1.ID,
		PaymentChannelID: pc1.ID,
	}
	if err := db.Create(&ppc1).Error; err != nil {
		return err
	}

	// 10. Subscriptions
	sub1 := subscriptions.Subscription{
		ID:          uuid.New(),
		Name:        "Mobile Money",
		Status:      "active",
		Description: "Give merchant access to Mobile Money transactions, collection and disbursement",
	}
	if err := db.Create(&sub1).Error; err != nil {
		return err
	}

	// 11. Merchant Fee Profiles (using a dummy MerchantID)
	mfp1 := merchantfeeprofiles.MerchantFeeProfile{
		ID:           uuid.New(),
		MerchantID:   uuid.New(),
		FeeProfileID: fp1.ID,
		Status:       "active",
	}
	if err := db.Create(&mfp1).Error; err != nil {
		return err
	}

	// 12. Merchant Subscriptions
	ms1 := subscriptions.MerchantSubscription{
		ID:             uuid.New(),
		MerchantID:     mfp1.MerchantID,
		SubscriptionID: sub1.ID,
		Status:         "active",
	}
	if err := db.Create(&ms1).Error; err != nil {
		return err
	}

	log.Println("Database seeded successfully!")
	return nil
}
