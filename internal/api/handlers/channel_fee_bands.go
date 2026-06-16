package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/channelfeebands"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/channel_fee_bands_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func CreateChannelFeeBandHandler(c *gin.Context) {
	var req channel_fee_bands_proto.CreateChannelFeeBandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := channelfeebands.CreateChannelFeeBand(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Channel fee bands created successfully")
}

func GetChannelFeeBandsHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	req := &channel_fee_bands_proto.GetChannelFeeBandsRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
	}

	if searchQuery != "" {
		req.SearchQuery = searchQuery
	}

	data, err := channelfeebands.GetChannelFeeBands(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Channel fee bands fetched successfully", gin.H{"data": data})
}

func GetChannelFeeBandHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	data, err := channelfeebands.GetChannelFeeBand(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Channel fee band fetched successfully", gin.H{"data": data})
}

func UpdateChannelFeeBandHandler(c *gin.Context) {
	var req channel_fee_bands_proto.EditChannelFeeBandsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := channelfeebands.UpdateChannelFeeBand(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Channel fee band updated successfully", nil)
}

func DeleteChannelFeeBandHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	err := channelfeebands.DeleteChannelFeeBand(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Channel fee band deleted successfully", nil)
}
