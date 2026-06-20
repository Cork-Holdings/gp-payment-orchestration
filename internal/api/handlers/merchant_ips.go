package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantips"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/merchant_ips_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func AddMerchantIPHandler(c *gin.Context) {
	var req merchant_ips_proto.CreateMerchantIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	err := merchantips.CreateMerchantIP(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func GetMerchantIPsHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")
	merchantID := c.Query("merchant_id")
	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)
	req := &merchant_ips_proto.GetMerchantIPsRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
		MerchantId:  merchantID,
	}
	data, err := merchantips.GetMerchantIPs(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant IPs fetched successfully", gin.H{"data": data})
}

func GetMerchantIPHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	data, err := merchantips.GetMerchantIP(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant IP fetched successfully", gin.H{"data": data})
}

func UpdateMerchantIPHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	var req merchant_ips_proto.EditMerchantIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	err := merchantips.UpdateMerchantIP(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant IP updated successfully", nil)
}

func DeleteMerchantIPHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	err := merchantips.DeleteMerchantIP(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant IP deleted successfully", nil)
}
