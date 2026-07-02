package cmd

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapikeys"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantfeeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantips"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantpaymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentservices"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/prefixes"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/providers"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/subscriptions"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/transactiontypes"
)

func RegisterAppModels(app *global.App) {
	if len(app.Models) > 0 {
		return
	}

	app.Register(
		&feeprofiles.FeeProfile{},
		&feeprofiles.ProfileFeeBands{},
		&providers.Provider{},
		&paymentservices.PaymentService{},
		&paymentchannels.PaymentChannel{},
		&paymentchannels.ChannelFeeBands{},
		&merchantpaymentchannels.MerchantPaymentChannel{},
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
}
