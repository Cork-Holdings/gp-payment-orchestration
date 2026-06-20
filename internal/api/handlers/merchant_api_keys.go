package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapikeys"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/merchant_api_keys_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
)

func CreateMerchantAPIKeyHandler() {

	merchantID := ""
	//listen on rabbitmq queue for merchantID
	err := global.GetMQ().Consume(global.New(), os.Getenv("QUEUE_NAME"), func(app *global.App, msg amqp091.Delivery) error {
		merchantID = string(msg.Body)
		return nil
	})
	if err != nil {
		log.Printf("Failed to consume merchantID: %v", err)
		return
	}

	//Generate API Keys
	merchant, err := merchantapikeys.CreateMerchantKeys(merchantID)
	if err != nil {
		log.Printf("Failed to create merchant API keys: %v", err)
		return
	}

	global.GetMQ().Emit("merchant_api_key.created", merchant)
}

func GetMerchantAPIKeysHandler(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	searchQuery := c.Query("search_query")
	merchantID := c.Query("merchant_id")

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	req := &merchant_api_keys_proto.GetMerchantAPIKeysRequest{
		Page:        int32(pageInt),
		PageSize:    int32(pageSizeInt),
		SearchQuery: searchQuery,
		MerchantId:  merchantID,
	}
	data, err := merchantapikeys.GetMerchantAPIKeys(req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant API keys fetched successfully", gin.H{"data": data})
}

func UpdateMerchantAPIKeyHandler(c *gin.Context) {

	var req merchant_api_keys_proto.EditMerchantAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	err := merchantapikeys.UpdateMerchantAPIKey(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant API keys updated successfully", nil)
}

func DeleteMerchantAPIKeyHandler(c *gin.Context) {
	id := c.Param("id")
	err := merchantapikeys.DeleteMerchantAPIKey(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Merchant API key deleted successfully", nil)
}

func GenerateAuthSignatureHandler(c *gin.Context) {
	var req merchant_api_keys_proto.GenerateAuthSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	signature, err := merchantapikeys.GenerateAuthSignature(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, "Auth signature generated successfully", gin.H{"signature": signature})
}
