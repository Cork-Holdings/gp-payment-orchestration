package routes

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(e *gin.Engine) {

	// group := e.Group("/fee-profiles")
	// {
	// 	group.POST("", handlers.CreateFeeProfile)
	// 	group.GET("", handlers.GetFeeProfiles)
	// 	group.GET("/:id", handlers.GetFeeProfile)
	// 	group.PUT("/:id", handlers.UpdateFeeProfile)
	// 	group.DELETE("/:id", handlers.DeleteFeeProfile)
	// }

	feeProfileRoutes := e.Group("/fee-profiles")
	{
		feeProfileRoutes.POST("/create", handlers.CreateFeeProfile)
		feeProfileRoutes.GET("/list", handlers.GetFeeProfiles)
		feeProfileRoutes.GET("/get/:id", handlers.GetFeeProfile)
		feeProfileRoutes.PUT("/update", handlers.UpdateFeeProfile)
		feeProfileRoutes.DELETE("/delete/:id", handlers.DeleteFeeProfile)
	}

	subscriptionRoutes := e.Group("/subscriptions")
	{
		subscriptionRoutes.POST("/create", handlers.CreateSubscription)
		subscriptionRoutes.GET("/list", handlers.GetSubscriptions)
		subscriptionRoutes.GET("/get/:id", handlers.GetSubscription)
		subscriptionRoutes.PUT("/update", handlers.UpdateSubscription)
		subscriptionRoutes.DELETE("/delete/:id", handlers.DeleteSubscription)
	}
}
