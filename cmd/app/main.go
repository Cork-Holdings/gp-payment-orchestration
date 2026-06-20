package main

import (
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/cmd"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapikeys"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantfeeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantips"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentservices"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/prefixes"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/providers"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/subscriptions"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/transactiontypes"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Failed to load env: %v", err.Error())
	}

	app := global.New()

	app.Register(
		&feeprofiles.FeeProfile{},
		&feeprofiles.ProfileFeeBands{},
		&providers.Provider{},
		&paymentservices.PaymentService{},
		&paymentchannels.PaymentChannel{},
		&paymentchannels.ChannelFeeBands{},
		&transactiontypes.TransactionType{},
		&transactiontypes.SubTransactionType{},
		&merchantfeeprofiles.MerchantFeeProfile{},
		&prefixes.Prefix{},
		&prefixes.PrefixPaymentChannel{},
		&subscriptions.Subscription{},
		&subscriptions.MerchantSubscription{},
		&merchantapikeys.MerchantAPIKey{},
		&merchantips.MerchantIP{},
	)

	cmd.Execute()
}
