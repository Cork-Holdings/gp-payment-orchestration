package main

import (
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/cmd"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/subscriptions"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Failed to load env: %v", err.Error())
	}

	app := global.New()

	app.Register(
		&feeprofiles.FeeProfile{},
		&feeprofiles.PaymentService{},
		&feeprofiles.PaymentChannel{},
		&feeprofiles.TransactionType{},
		&feeprofiles.SubTransactionType{},
		&feeprofiles.MerchantFeeProfile{},
		&feeprofiles.ChannelFeeBands{},
		&feeprofiles.ProfileFeeBands{},
		&feeprofiles.Prefix{},
		&feeprofiles.PrefixPaymentChannel{},
		&subscriptions.Subscription{},
		&subscriptions.MerchantSubscription{},
	)

	cmd.Execute()
}
