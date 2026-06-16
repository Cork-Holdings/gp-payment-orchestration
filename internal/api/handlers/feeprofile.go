package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/fee_profiles_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func NewFeeProfileHandler(c *gin.Context) {
	var req fee_profiles_proto.CreateFeeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := feeprofiles.CreateFeeProfile(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Fee profile created successfully", nil)
}

func GetFeeProfilesHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	req := &fee_profiles_proto.GetFeeProfilesRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
	}

	data, err := feeprofiles.GetFeeProfiles(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Fee profiles fetched successfully", gin.H{"data": data})
}

func GetFeeProfileHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	data, err := feeprofiles.GetFeeProfile(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Fee profile fetched successfully", gin.H{"data": data})
}

func UpdateFeeProfileHandler(c *gin.Context) {
	var req fee_profiles_proto.EditFeeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := feeprofiles.UpdateFeeProfile(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Fee profile updated successfully", nil)
}

func DeleteFeeProfileHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "ID is required")
		return
	}
	err := feeprofiles.DeleteFeeProfile(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Fee profile deleted successfully", nil)
}

// func ApproveFeeProfile(c *gin.Context) {
// 	id := c.Param("id")
// 	userID := c.GetString("user_id") // Assuming user_id is in context from auth middleware

// 	err := feeprofiles.ApproveFeeProfile(&req)
// 	if err != nil {
// 		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	utils.RespondWithSuccess(c, "Fee profile approved successfully", gin.H{"data": data})
// }

// func RejectFeeProfile(c *gin.Context) {
// 	id := c.Param("id")
// 	userID := c.GetString("user_id")
// 	var req struct {
// 		Reason string `json:"reason"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	updates := map[string]any{
// 		"approval_status": "rejected",
// 		"rejected_by":     uuid.MustParse(userID),
// 		"rejected_at":     time.Now(),
// 		"rejected_reason": req.Reason,
// 	}

// 	data, err := h.service.UpdateFeeProfile(id, updates)
// 	if err != nil {
// 		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	utils.RespondWithSuccess(c, "Fee profile rejected successfully", gin.H{"data": data})
// }
