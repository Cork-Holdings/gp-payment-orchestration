package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/transactiontypes"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/transaction_types_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
)

func CreateTransactionTypeHandler(c *gin.Context) {
	var req transaction_types_proto.CreateTransactionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := transactiontypes.CreateTransactionType(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Transaction type created successfully")
}

func GetTransactionTypesHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	req := &transaction_types_proto.GetTransactionTypesRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
	}
	data, err := transactiontypes.GetTransactionTypes(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Transaction types fetched successfully", gin.H{"data": data})
}

func GetTransactionTypeHandler(c *gin.Context) {
	id := c.Param("id")
	data, err := transactiontypes.GetTransactionType(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Transaction type fetched successfully", gin.H{"data": data})
}

func UpdateTransactionTypeHandler(c *gin.Context) {
	var req transaction_types_proto.EditTransactionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := transactiontypes.UpdateTransactionType(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Transaction type updated successfully")
}

func DeleteTransactionTypeHandler(c *gin.Context) {
	id := c.Param("id")
	err := transactiontypes.DeleteTransactionType(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Transaction type deleted successfully", nil)
}

// SubTransactionType Handlers

func CreateSubTransactionTypeHandler(c *gin.Context) {
	var req transaction_types_proto.CreateSubTransactionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := transactiontypes.CreateSubTransactionType(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Sub transaction type created successfully")
}

func GetSubTransactionTypesHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")
	transactionTypeID := c.Query("transaction_type_id")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	data, err := transactiontypes.GetSubTransactionTypes(&transaction_types_proto.GetSubTransactionTypesRequest{
		Page:              int32(pageInt),
		PageSize:          int32(pageSizeInt),
		SearchQuery:       searchQuery,
		TransactionTypeId: transactionTypeID,
	})
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Sub transaction types fetched successfully", gin.H{"data": data})
}

func GetSubTransactionTypeHandler(c *gin.Context) {
	id := c.Param("id")
	data, err := transactiontypes.GetSubTransactionType(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Sub transaction type fetched successfully", gin.H{"data": data})
}

func UpdateSubTransactionTypeHandler(c *gin.Context) {
	var req transaction_types_proto.EditSubTransactionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := transactiontypes.UpdateSubTransactionType(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(c, "Sub transaction type updated successfully")
}

func DeleteSubTransactionTypeHandler(c *gin.Context) {
	id := c.Param("id")
	err := transactiontypes.DeleteSubTransactionType(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Sub transaction type deleted successfully", nil)
}
