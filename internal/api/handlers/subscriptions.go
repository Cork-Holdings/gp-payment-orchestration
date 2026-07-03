package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/subscriptions"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/subscriptions_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func CreateSubscriptionHandler(c *gin.Context) {

	var req subscriptions_proto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// utils.Log(slog.Error, err.Error())
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := subscriptions.CreateSubscription(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Subscription created successfully")

}

func GetSubscriptionsHandler(c *gin.Context) {

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid page")
		return
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid page size")
		return
	}

	req := &subscriptions_proto.GetSubscriptionsRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
	}

	data, err := subscriptions.GetSubscriptions(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Subscriptions fetched successfully", gin.H{
		"data": data,
	})
}

func GetSubscriptionHandler(c *gin.Context) {
}

func UpdateSubscriptionHandler(c *gin.Context) {

	var req subscriptions_proto.EditSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// utils.Log(slog.Error, err.Error())
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := subscriptions.UpdateSubscription(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Subscription updated successfully")

}

func DeleteSubscriptionHandler(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Subscription ID is required")
		return
	}

	err := subscriptions.DeleteSubscription(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func CreateMerchantSubscriptionHandler(c *gin.Context) {
	var req subscriptions_proto.CreateMerchantSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := subscriptions.CreateMerchantSubscription(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Merchant subscription created successfully")
}

func GetMerchantSubscriptionsHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid page")
		return
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid page size")
		return
	}

	req := &subscriptions_proto.GetMerchantSubscriptionsRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
	}

	data, err := subscriptions.GetMerchantSubscriptions(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Merchant subscriptions fetched successfully", gin.H{
		"data": data,
	})
}

// func GetMerchantSubscription(c *gin.Context) {
// 	id := c.Param("id")

// 	if id == "" {
// 		utils.RespondWithError(c, http.StatusBadRequest, "Merchant subscription ID is required")
// 		return
// 	}

// 	err := subscriptions.GetMerchantSubscription(id)
// 	if err != nil {
// 		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
// 		return
// 	}
// 	utils.RespondWithSuccess(c, "Merchant subscription fetched successfully", gin.H{
// 		"data": data,
// 	})
// }

func UpdateMerchantSubscriptionHandler(c *gin.Context) {
	var req subscriptions_proto.EditMerchantSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := subscriptions.UpdateMerchantSubscription(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant subscription updated successfully")
}

func DeleteMerchantSubscriptionHandler(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Merchant subscription ID is required")
		return
	}

	err := subscriptions.DeleteMerchantSubscription(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant subscription deleted successfully")
}
