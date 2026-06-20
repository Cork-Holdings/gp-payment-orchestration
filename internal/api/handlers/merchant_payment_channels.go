package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantpaymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/merchant_payment_channels_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func CreateMerchantPaymentChannelHandler(c *gin.Context) {
	var req merchant_payment_channels_proto.CreateMerchantPaymentChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	err := merchantpaymentchannels.CreateMerchantPaymentChannel(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant payment channel created successfully")
}

func GetMerchantPaymentChannelsHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")
	merchantId := c.Query("merchant_id")
	paymentChannelId := c.Query("payment_channel_id")
	status := c.Query("status")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	req := &merchant_payment_channels_proto.GetMerchantPaymentChannelsRequest{
		Page:             int32(pageInt),
		PageSize:         int32(pageSizeInt),
		SearchQuery:      searchQuery,
		MerchantId:       merchantId,
		PaymentChannelId: paymentChannelId,
		Status:           status,
	}

	data, err := merchantpaymentchannels.GetMerchantPaymentChannels(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Merchant payment channels fetched successfully", gin.H{"data": data})
}

func GetMerchantPaymentChannelHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	data, err := merchantpaymentchannels.GetMerchantPaymentChannel(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant payment channel fetched successfully", gin.H{"data": data})
}

func UpdateMerchantPaymentChannelHandler(c *gin.Context) {
	var req merchant_payment_channels_proto.EditMerchantPaymentChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	err := merchantpaymentchannels.UpdateMerchantPaymentChannel(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant payment channel updated successfully")
}

func DeleteMerchantPaymentChannelHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	err := merchantpaymentchannels.DeleteMerchantPaymentChannel(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant payment channel deleted successfully")
}
