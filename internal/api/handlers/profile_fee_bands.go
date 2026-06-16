package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/profilefeebands"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/profile_fee_bands_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func CreateProfileFeeBandsHandler(c *gin.Context) {
	var req profile_fee_bands_proto.CreateProfileFeeBandsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := profilefeebands.CreateProfileFeeBands(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Profile fee bands created successfully")
}

func GetProfileFeeBandsHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	req := &profile_fee_bands_proto.GetProfileFeeBandsRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
	}

	data, err := profilefeebands.GetProfileFeeBands(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Profile fee bands fetched successfully", gin.H{"data": data})
}

func GetProfileFeeBandHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	data, err := profilefeebands.GetProfileFeeBand(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Profile fee band fetched successfully", gin.H{"data": data})
}

func UpdateProfileFeeBandHandler(c *gin.Context) {
	var req profile_fee_bands_proto.EditProfileFeeBandsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := profilefeebands.UpdateProfileFeeBand(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Profile fee band updated successfully")
}

func DeleteProfileFeeBandHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	err := profilefeebands.DeleteProfileFeeBand(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Profile fee band deleted successfully", nil)
}
