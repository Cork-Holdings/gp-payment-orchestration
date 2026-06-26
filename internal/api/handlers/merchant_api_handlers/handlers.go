package merchantapihandlers

import (
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

	clientID := c.GetString("client_id")

	collectReq := &merchantapis.CollectRequest{
		ClientID:       clientID,
		PhoneNumber:    req.PhoneNumber,
		Amount:         req.Amount,
		TransactionRef: transactionRef,
	}

	resp, err := merchantapis.HandleCollection(global.New(), collectReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	clientID := c.GetString("client_id")

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

	clientID := c.GetString("client_id")

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

	clientID := c.GetString("client_id")

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

	clientID := c.GetString("client_id")

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

	clientID := c.GetString("client_id")

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
