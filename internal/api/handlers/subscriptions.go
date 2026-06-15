package handlers

import (
	"net/http"

	subscriptionsservice "github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/subscriptions_Service"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/subscriptions_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func CreateSubscription(c *gin.Context) {

	var req subscriptions_proto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// utils.Log(slog.Error, err.Error())
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := subscriptionsservice.CreateSubscription(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Subscription created successfully")

}

func GetSubscriptions(c *gin.Context) {
}

func GetSubscription(c *gin.Context) {
}

func UpdateSubscription(c *gin.Context) {

}

func DeleteSubscription(c *gin.Context) {
}
