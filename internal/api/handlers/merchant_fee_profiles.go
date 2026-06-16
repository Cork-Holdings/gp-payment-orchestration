package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantfeeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/merchant_fee_profile_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func CreateMerchantFeeProfileHandler(c *gin.Context) {
	var req merchant_fee_profile_proto.CreateMerchantFeeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := merchantfeeprofiles.CreateMerchantFeeProfile(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Merchant fee profile created successfully")
}

func GetMerchantFeeProfilesHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	req := &merchant_fee_profile_proto.GetMerchantFeeProfilesRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
	}

	data, err := merchantfeeprofiles.GetMerchantFeeProfiles(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Merchant fee profiles fetched successfully", gin.H{"data": data})
}

func GetMerchantFeeProfileHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	data, err := merchantfeeprofiles.GetMerchantFeeProfile(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant fee profile fetched successfully", gin.H{"data": data})
}

func UpdateMerchantFeeProfileHandler(c *gin.Context) {
	var req merchant_fee_profile_proto.EditMerchantFeeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := merchantfeeprofiles.UpdateMerchantFeeProfile(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Merchant fee profile updated successfully", nil)
}

func DeleteMerchantFeeProfileHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	err := merchantfeeprofiles.DeleteMerchantFeeProfile(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant fee profile deleted successfully", nil)
}
