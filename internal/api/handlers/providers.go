package handlers

import (
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/providers"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/providers_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func CreateProviderHandler(c *gin.Context) {

	var req providers_proto.CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, 400, "Unable to bind Json", err.Error())
		return
	}

	if err := providers.CreateProvider(&req); err != nil {
		utils.RespondWithError(c, 400, "Failed to create provider", err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Provider created successfully")

}

func GetProvidersHandler(c *gin.Context) {

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	providerName := c.DefaultQuery("provider_name", "")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid page parameter", err.Error())
		return

	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid page_size parameter", err.Error())
		return
	}

	req := &providers_proto.GetProvidersRequest{
		Page:         int32(pageInt),
		PageSize:     int32(pageSizeInt),
		ProviderName: providerName,
	}

	providers, err := providers.GetProviders(req)
	if err != nil {
		utils.RespondWithError(c, 400, "Failed to get providers", err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Providers fetched successfully", gin.H{
		"data": providers,
	})

}

func UpdateProviderHandler(c *gin.Context) {
	var req providers_proto.EditProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, 400, "Unable to bind Json", err.Error())
		return
	}

	if err := providers.UpdateProvider(&req); err != nil {
		utils.RespondWithError(c, 400, "Failed to update provider", err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Provider updated successfully")

}

func DeleteProviderHandler(c *gin.Context) {

	providerId := c.Param("id")
	if providerId == "" {
		utils.RespondWithError(c, 400, "Provider ID is required", "Missing provider ID")
		return
	}

	err := providers.DeleteProvider(providerId)
	if err != nil {
		utils.RespondWithError(c, 400, "Failed to delete provider", err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Provider deleted successfully")

}

func GetProviderHandler(c *gin.Context) {

	providerId := c.Param("id")
	if providerId == "" {
		utils.RespondWithError(c, 400, "Provider ID is required", "Missing provider ID")
		return
	}

	provider, err := providers.GetProvider(providerId)
	if err != nil {
		utils.RespondWithError(c, 400, "Failed to get provider", err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Provider fetched successfully", gin.H{
		"data": provider,
	})

}
