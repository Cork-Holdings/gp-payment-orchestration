package merchantapihandlers

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapis"
	"github.com/gin-gonic/gin"
)

func HandleGenerateTokenHandler(c *gin.Context) {
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	if clientID == "" || clientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing client_id or client_secret"})
		return
	}

	app := global.New()

	req := merchantapis.TokenRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

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

	// // Calculate MNO fees
	// fee := merchantapis.CalculateMNOfees(req.PhoneNumber, req.Amount)

	collectReq := &merchantapis.CollectRequest{
		ClientID:       clientID,
		PhoneNumber:    req.PhoneNumber,
		Amount:         req.Amount, // Include fee in the collection amount
		TransactionRef: transactionRef,
	}
	err := merchantapis.HandleCollect(global.New(), collectReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collection initiated successfully"})
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

	rawBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	var req struct {
		PhoneNumber string  `json:"phone_number" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		// Currency    string  `json:"currency" binding:"required"`
	}

	if err := json.Unmarshal(rawBody, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body: " + err.Error()})
		return
	}

	clientID := c.GetString("client_id")

	disburseReq := &merchantapis.DisburseRequest{
		ClientID:       clientID,
		PhoneNumber:    req.PhoneNumber,
		Amount:         req.Amount,
		Signature:      signature,
		RawBody:        string(rawBody),
		TransactionRef: transactionRef,
	}

	resp, err := merchantapis.HandleDisburse(app, disburseReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
