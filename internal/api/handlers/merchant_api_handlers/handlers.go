package merchantapihandlers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapis"
	"github.com/gin-gonic/gin"
)

func HandleGenerateTokenHandler(c *gin.Context) {
	var req merchantapis.TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	fmt.Println("=============================================================")
	fmt.Println("Request: ", req)
	fmt.Println("=============================================================")

	if req.ClientID == "" || req.ClientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing client_id or client_secret"})
		return
	}

	app := global.New()

	resp, err := merchantapis.HandleGenerateToken(app, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication request failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)

}

var phoneValidationRegex = regexp.MustCompile(`^\d{12}$`)

func HandleCollectionHandler(c *gin.Context) {
	// app := global.New()

	//Check for X-Transaction-Ref Header
	transactionRef := c.GetHeader("X-Transaction-Ref")
	if transactionRef == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Transaction-Ref header"})
		return
	}

	var req struct {
		PhoneNumber string  `json:"phone_number" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		// Currency    string  `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate phone_number is exactly 12 digits
	if !phoneValidationRegex.MatchString(req.PhoneNumber) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone_number must be exactly 12 digits long"})
		return
	}

	clientID := c.GetHeader("X-Client-ID")

	collectReq := &merchantapis.CollectRequest{
		ClientID:       clientID,
		PhoneNumber:    req.PhoneNumber,
		Amount:         req.Amount,
		TransactionRef: transactionRef,
	}

	resp := merchantapis.HandleCollection(global.New(), collectReq)

	c.JSON(resp.Code, resp)
}

// B. Mobile Money Disbursements (Synchronous, X-Auth-Signature validated)
func HandleDisbursementHandler(c *gin.Context) {
	app := global.New()

	//Check for X-Transaction-Ref Header
	transactionRef := c.GetHeader("X-Transaction-Ref")
	if transactionRef == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Transaction-Ref header"})
		return
	}

	signature := c.GetHeader("X-Auth-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Auth-Signature header"})
		return
	}

	var req struct {
		PhoneNumber string  `json:"phone_number" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		Narration   string  `json:"narration"`
		// Currency    string  `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body: " + err.Error()})
		return
	}

	clientID := c.GetHeader("X-Client-ID")

	disburseReq := &merchantapis.DisburseRequest{
		ClientID:       clientID,
		PhoneNumber:    req.PhoneNumber,
		Amount:         req.Amount,
		Narration:      req.Narration,
		Signature:      signature,
		TransactionRef: transactionRef,
	}

	resp, err := merchantapis.HandleDisbursement(app, disburseReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func HandleCollectionCheckStatusHandler(c *gin.Context) {

	transactionRef := c.Param("transaction_ref")

	if transactionRef == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "missing transaction_ref"})
		return
	}

	clientID := c.GetHeader("X-Client-ID")

	checkStatusReq := &merchantapis.CheckStatusRequest{
		TransactionRef: transactionRef,
		ClientID:       clientID,
	}

	checkStatusResp, err := merchantapis.HandleCollectionCheckStatus(global.New(), checkStatusReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(checkStatusResp.Code, checkStatusResp)
}

func HandleCollectionCheckBalanceHandler(c *gin.Context) {

	clientID := c.GetHeader("X-Client-ID")

	checkBalanceReq := &merchantapis.CheckCollectionBalanceRequest{
		ClientID: clientID,
	}

	checkBalanceResp, err := merchantapis.HandleCollectionCheckBalance(global.New(), checkBalanceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(checkBalanceResp.Code, checkBalanceResp)

}

func HandleDisbursementCheckStatusHandler(c *gin.Context) {

	transactionRef := c.Param("transaction_ref")

	if transactionRef == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "missing transaction_ref"})
		return
	}

	clientID := c.GetHeader("X-Client-ID")

	checkStatusReq := &merchantapis.CheckStatusRequest{
		TransactionRef: transactionRef,
		ClientID:       clientID,
	}

	checkStatusResp, err := merchantapis.HandleDisbursementCheckStatus(global.New(), checkStatusReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(checkStatusResp.Code, checkStatusResp)

}

func HandleDisbursementCheckBalanceHandler(c *gin.Context) {

	clientID := c.GetHeader("X-Client-ID")

	xAuthSignature := c.GetHeader("X-Auth-Signature")
	if xAuthSignature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Auth-Signature header"})
		return
	}

	checkBalanceReq := &merchantapis.CheckDisbursementBalanceRequest{
		ClientID:       clientID,
		XAuthSignature: xAuthSignature,
	}

	checkBalanceResp, err := merchantapis.HandleDisbursementCheckBalance(global.New(), checkBalanceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(checkBalanceResp.Code, checkBalanceResp)

}

func HandleCreateCheckoutSessionHandler(c *gin.Context) {

}

// func HandleCreateDummyMerchantHandler(c *gin.Context) {
// 	merchantID := uuid.NewString()

// 	merchantKey, err := merchantapikeys.CreateMerchantKeys(merchantID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"message":       "Dummy merchant created successfully",
// 		"merchant_id":   merchantID,
// 		"client_id":     merchantKey.ClientID,
// 		"client_secret": merchantKey.ClientSecret,
// 	})
// }

// func HandleGetMerchantAccountsHandler(c *gin.Context) {
// 	app := global.New()
// 	clientID := c.GetString("client_id")

// 	var key merchantapikeys.MerchantAPIKey
// 	if err := app.DB.Where("client_id = ?", clientID).First(&key).Error; err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to resolve merchant from client_id"})
// 		return
// 	}

// 	rpcResp, err := app.MQ.Request("merchant.accounts.get", map[string]any{
// 		"merchant_id": key.MerchantID.String(),
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var payload map[string]any
// 	if err := json.Unmarshal(rpcResp, &payload); err != nil {
// 		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid response from transactions service"})
// 		return
// 	}

// 	code := http.StatusOK
// 	if rawCode, ok := payload["code"].(float64); ok {
// 		code = int(rawCode)
// 	}

// 	c.JSON(code, payload)
// }

// func HandleGetMerchantAccountTransactionsHandler(c *gin.Context) {
// 	app := global.New()
// 	clientID := c.GetString("client_id")

// 	var key merchantapikeys.MerchantAPIKey
// 	if err := app.DB.Where("client_id = ?", clientID).First(&key).Error; err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to resolve merchant from client_id"})
// 		return
// 	}

// 	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
// 	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

// 	payload := map[string]any{
// 		"merchant_id":      key.MerchantID.String(),
// 		"page":             page,
// 		"page_size":        pageSize,
// 		"transaction_id":   c.Query("transaction_id"),
// 		"transaction_type": c.Query("transaction_type"),
// 		"start_date":       c.Query("start_date"),
// 		"end_date":         c.Query("end_date"),
// 	}

// 	rpcResp, err := app.MQ.Request("merchant.accounts.transactions.get", payload)
// 	if err != nil {
// 		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var rpcPayload map[string]any
// 	if err := json.Unmarshal(rpcResp, &rpcPayload); err != nil {
// 		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid response from transactions service"})
// 		return
// 	}

// 	code := http.StatusOK
// 	if rawCode, ok := rpcPayload["code"].(float64); ok {
// 		code = int(rawCode)
// 	}

// 	c.JSON(code, rpcPayload)
// }

func HandleNameLookupHandler(c *gin.Context) {

	clientID := c.GetHeader("X-Client-ID")

	phone := c.Param("phone")

	nameLookupReq := &merchantapis.NameLookupRequest{
		ClientID:    clientID,
		PhoneNumber: phone,
	}

	nameLookupResp, err := merchantapis.HandleNameLookup(global.New(), nameLookupReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, nameLookupResp)

}
