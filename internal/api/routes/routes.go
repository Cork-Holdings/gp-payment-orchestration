package routes

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/api/handlers"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(e *gin.Engine, app *global.App) {

	feeProfileRoutes := e.Group("/fee-profiles")
	{
		feeProfileRoutes.POST("/create", handlers.NewFeeProfileHandler)
		feeProfileRoutes.GET("/list", handlers.GetFeeProfilesHandler)
		feeProfileRoutes.GET("/get/:id", handlers.GetFeeProfileHandler)
		feeProfileRoutes.PUT("/update", handlers.UpdateFeeProfileHandler)
		feeProfileRoutes.DELETE("/delete/:id", handlers.DeleteFeeProfileHandler)
	}

	paymentServiceRoutes := e.Group("/payment-services")
	{
		paymentServiceRoutes.POST("/create", handlers.CreatePaymentServiceHandler)
		paymentServiceRoutes.GET("/list", handlers.GetPaymentServicesHandler)
		paymentServiceRoutes.GET("/get/:id", handlers.GetPaymentServiceHandler)
		paymentServiceRoutes.PUT("/update", handlers.UpdatePaymentServiceHandler)
		paymentServiceRoutes.DELETE("/delete/:id", handlers.DeletePaymentServiceHandler)
	}

	paymentChannelRoutes := e.Group("/payment-channels")
	{
		paymentChannelRoutes.POST("/create", handlers.CreatePaymentChannelHandler)
		paymentChannelRoutes.GET("/list", handlers.GetPaymentChannelsHandler)
		paymentChannelRoutes.GET("/get/:id", handlers.GetPaymentChannelHandler)
		paymentChannelRoutes.PUT("/update", handlers.UpdatePaymentChannelHandler)
		paymentChannelRoutes.DELETE("/delete/:id", handlers.DeletePaymentChannelHandler)
	}

	transactionTypeRoutes := e.Group("/transaction-types")
	{
		transactionTypeRoutes.POST("/create", handlers.CreateTransactionTypeHandler)
		transactionTypeRoutes.GET("/list", handlers.GetTransactionTypesHandler)
		transactionTypeRoutes.GET("/get/:id", handlers.GetTransactionTypeHandler)
		transactionTypeRoutes.PUT("/update", handlers.UpdateTransactionTypeHandler)
		transactionTypeRoutes.DELETE("/delete/:id", handlers.DeleteTransactionTypeHandler)

		// SubTransactionTypes
		transactionTypeRoutes.POST("/sub/create", handlers.CreateSubTransactionTypeHandler)
		transactionTypeRoutes.GET("/sub/list", handlers.GetSubTransactionTypesHandler)
		transactionTypeRoutes.GET("/sub/get/:id", handlers.GetSubTransactionTypeHandler)
		transactionTypeRoutes.PUT("/sub/update", handlers.UpdateSubTransactionTypeHandler)
		transactionTypeRoutes.DELETE("/sub/delete/:id", handlers.DeleteSubTransactionTypeHandler)
	}

	merchantFeeProfileRoutes := e.Group("/merchant-fee-profiles")
	{
		merchantFeeProfileRoutes.POST("/create", handlers.CreateMerchantFeeProfileHandler)
		merchantFeeProfileRoutes.GET("/list", handlers.GetMerchantFeeProfilesHandler)
		merchantFeeProfileRoutes.GET("/get/:id", handlers.GetMerchantFeeProfileHandler)
		merchantFeeProfileRoutes.PUT("/update", handlers.UpdateMerchantFeeProfileHandler)
		merchantFeeProfileRoutes.DELETE("/delete/:id", handlers.DeleteMerchantFeeProfileHandler)
	}

	channelFeeBandsRoutes := e.Group("/channel-fee-bands")
	{
		channelFeeBandsRoutes.POST("/create", handlers.CreateChannelFeeBandHandler)
		channelFeeBandsRoutes.GET("/list", handlers.GetChannelFeeBandsHandler)
		channelFeeBandsRoutes.GET("/get/:id", handlers.GetChannelFeeBandHandler)
		channelFeeBandsRoutes.PUT("/update", handlers.UpdateChannelFeeBandHandler)
		channelFeeBandsRoutes.DELETE("/delete/:id", handlers.DeleteChannelFeeBandHandler)
	}

	profileFeeBandsRoutes := e.Group("/profile-fee-bands")
	{
		profileFeeBandsRoutes.POST("/create", handlers.CreateProfileFeeBandsHandler)
		profileFeeBandsRoutes.GET("/list", handlers.GetProfileFeeBandsHandler)
		profileFeeBandsRoutes.GET("/get/:id", handlers.GetProfileFeeBandHandler)
		profileFeeBandsRoutes.PUT("/update", handlers.UpdateProfileFeeBandHandler)
		profileFeeBandsRoutes.DELETE("/delete/:id", handlers.DeleteProfileFeeBandHandler)
	}

	prefixesRoutes := e.Group("/prefixes")
	{
		prefixesRoutes.POST("/create", handlers.CreatePrefixHandler)
		prefixesRoutes.GET("/list", handlers.GetPrefixesHandler)
		prefixesRoutes.GET("/get/:id", handlers.GetPrefixHandler)
		prefixesRoutes.PUT("/update", handlers.UpdatePrefixHandler)
		prefixesRoutes.DELETE("/delete/:id", handlers.DeletePrefixHandler)

		// PrefixPaymentChannels
		prefixesRoutes.POST("/channels/create", handlers.CreatePrefixPaymentChannelHandler)
		prefixesRoutes.GET("/channels/list", handlers.GetPrefixPaymentChannelsHandler)
		prefixesRoutes.DELETE("/channels/delete/:id", handlers.DeletePrefixPaymentChannelHandler)
	}

	subscriptionRoutes := e.Group("/subscriptions")
	{
		subscriptionRoutes.POST("/create", handlers.CreateSubscriptionHandler)
		subscriptionRoutes.GET("/list", handlers.GetSubscriptionsHandler)
		subscriptionRoutes.GET("/get/:id", handlers.GetSubscriptionHandler)
		subscriptionRoutes.PUT("/update", handlers.UpdateSubscriptionHandler)
		subscriptionRoutes.DELETE("/delete/:id", handlers.DeleteSubscriptionHandler)
	}

	merchantSubscriptionRoutes := e.Group("/merchant-subscriptions")
	{
		merchantSubscriptionRoutes.POST("/create", handlers.CreateMerchantSubscriptionHandler)
		merchantSubscriptionRoutes.GET("/list", handlers.GetMerchantSubscriptionsHandler)
		merchantSubscriptionRoutes.PUT("/update", handlers.UpdateMerchantSubscriptionHandler)
		merchantSubscriptionRoutes.DELETE("/delete/:id", handlers.DeleteMerchantSubscriptionHandler)
	}
}
