package routes

import (
	"context"
	"net/http"
	"os"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/proto/ledgerpb"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func RegisterLedgerRoutes(e *gin.Engine, app *global.App) {
	// Establish gRPC connection to Ledger Service
	addr := os.Getenv("LEDGER_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50052"
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to ledger service: " + err.Error())
	}

	client := ledgerpb.NewLedgerServiceClient(conn)

	e.POST("/ledger/accounts", func(c *gin.Context) {
		var req struct {
			Name           string  `json:"name" binding:"required"`
			Currency       string  `json:"currency" binding:"required"`
			InitialBalance float64 `json:"initial_balance"`
			WalletType     string  `json:"wallet_type"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := client.CreateAccount(context.Background(), &ledgerpb.CreateAccountRequest{
			Name:           req.Name,
			Currency:       req.Currency,
			InitialBalance: req.InitialBalance,
			WalletType:     req.WalletType,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	e.GET("/ledger/accounts/:id/balance", func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account ID is required"})
			return
		}

		res, err := client.GetBalance(context.Background(), &ledgerpb.GetBalanceRequest{
			AccountId: id,
		})
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	e.POST("/ledger/transfers", func(c *gin.Context) {
		var req struct {
			SourceAccountId      string  `json:"source_account_id" binding:"required"`
			DestinationAccountId string  `json:"destination_account_id" binding:"required"`
			Amount               float64 `json:"amount" binding:"required,gt=0"`
			Currency             string  `json:"currency" binding:"required"`
			TransferId           string  `json:"transfer_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := client.Transfer(context.Background(), &ledgerpb.TransferRequest{
			SourceAccountId:      req.SourceAccountId,
			DestinationAccountId: req.DestinationAccountId,
			Amount:               req.Amount,
			Currency:             req.Currency,
			TransferId:           req.TransferId,
		})
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.InvalidArgument && st.Message() == "CROSS_WALLET_NOT_ALLOWED" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": st.Message(),
				})
				return
			}
			if ok && st.Code() == codes.Aborted {
				c.JSON(http.StatusConflict, gin.H{
					"error": "CONCURRENT_MODIFICATION_ERROR",
					"message": st.Message(),
				})
				return
			}
			if ok && st.Code() == codes.FailedPrecondition {
				c.JSON(http.StatusPaymentRequired, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	e.POST("/ledger/transfers/:id/status", func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := client.UpdateTransferStatus(context.Background(), &ledgerpb.UpdateStatusRequest{
			TransferId: id,
			NewStatus:  req.Status,
		})
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.FailedPrecondition && st.Message() == "INVALID_TRANSFER_STATE" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": st.Message(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	e.POST("/ledger/batches", func(c *gin.Context) {
		var req struct {
			Transfers []struct {
				SourceAccountId      string  `json:"source_account_id" binding:"required"`
				DestinationAccountId string  `json:"destination_account_id" binding:"required"`
				Amount               float64 `json:"amount" binding:"required,gt=0"`
				Currency             string  `json:"currency" binding:"required"`
				TransferId           string  `json:"transfer_id"`
			} `json:"transfers" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(req.Transfers) > 10000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bulk distributions capped at 10,000 transfers maximum"})
			return
		}

		var pbTransfers []*ledgerpb.TransferRequest
		for _, t := range req.Transfers {
			pbTransfers = append(pbTransfers, &ledgerpb.TransferRequest{
				SourceAccountId:      t.SourceAccountId,
				DestinationAccountId: t.DestinationAccountId,
				Amount:               t.Amount,
				Currency:             t.Currency,
				TransferId:           t.TransferId,
			})
		}

		res, err := client.BatchTransfer(context.Background(), &ledgerpb.BatchTransferRequest{
			Transfers: pbTransfers,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})
}
