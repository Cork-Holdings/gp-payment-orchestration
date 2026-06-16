package grpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	ledger "github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/ledger"
	"github.com/Cork-Holdings/gp_payment_orchestration/proto/ledgerpb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type LedgerServer struct {
	ledgerpb.UnimplementedLedgerServiceServer
	app *global.App
}

func NewLedgerServer(app *global.App) *LedgerServer {
	return &LedgerServer{app: app}
}

func (s *LedgerServer) CreateAccount(ctx context.Context, req *ledgerpb.CreateAccountRequest) (*ledgerpb.AccountResponse, error) {
	walletType := req.WalletType
	if walletType == "" {
		walletType = "emoney"
	}
	acc := &ledger.Account{
		Name:       req.Name,
		Currency:   req.Currency,
		Balance:    req.InitialBalance,
		Version:    1,
		WalletType: walletType,
	}
	acc.Autofill(acc)

	err := s.app.DB.Create(acc).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create account: %v", err)
	}

	return &ledgerpb.AccountResponse{
		AccountId:  *acc.ExtID,
		Name:       acc.Name,
		Currency:   acc.Currency,
		Balance:    acc.Balance,
		Version:    acc.Version,
		WalletType: acc.WalletType,
	}, nil
}

func (s *LedgerServer) GetBalance(ctx context.Context, req *ledgerpb.GetBalanceRequest) (*ledgerpb.BalanceResponse, error) {
	cacheKey := fmt.Sprintf("balance:%s", req.AccountId)
	
	// Check Redis cache first
	val, err := s.app.Cache.Get(ctx, cacheKey).Float64()
	if err == nil {
		// Cache hit
		return &ledgerpb.BalanceResponse{
			AccountId: req.AccountId,
			Balance:   val,
			Currency:  "USD", // Mock or retrieve from metadata
			Cached:    true,
		}, nil
	}

	// Cache miss, check PostgreSQL GORM database
	var acc ledger.Account
	err = s.app.DB.Where("ext_id = ?", req.AccountId).First(&acc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	// Store in cache (5 minutes TTL)
	s.app.Cache.Set(ctx, cacheKey, acc.Balance, 5*time.Minute)

	return &ledgerpb.BalanceResponse{
		AccountId: *acc.ExtID,
		Balance:   acc.Balance,
		Currency:  acc.Currency,
		Cached:    false,
	}, nil
}

func (s *LedgerServer) Transfer(ctx context.Context, req *ledgerpb.TransferRequest) (*ledgerpb.TransferResponse, error) {
	transferID := req.TransferId
	if transferID == "" {
		transferID = "txn_" + time.Now().Format("20060102150405")
	}

	// 1. Check that source and destination exist
	var src ledger.Account
	var dest ledger.Account
	if err := s.app.DB.Where("ext_id = ?", req.SourceAccountId).First(&src).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "source account not found")
	}
	if err := s.app.DB.Where("ext_id = ?", req.DestinationAccountId).First(&dest).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "destination account not found")
	}

	// Step 3: Inject Wallet-Type validation check (return CROSS_WALLET_NOT_ALLOWED)
	if src.WalletType != dest.WalletType {
		return nil, status.Errorf(codes.InvalidArgument, "CROSS_WALLET_NOT_ALLOWED")
	}

	// Step 2: Push atomic WAL entry to Redis Stream before execution
	walVal := map[string]interface{}{
		"transfer_id":   transferID,
		"source_id":     req.SourceAccountId,
		"dest_id":       req.DestinationAccountId,
		"amount":        req.Amount,
		"currency":      req.Currency,
		"timestamp":     time.Now().Format(time.RFC3339),
	}
	err := s.app.Cache.XAdd(ctx, &redis.XAddArgs{
		Stream: "ledger:wal",
		Values: walVal,
	}).Err()
	if err != nil {
		log.Printf("[WAL] Redis Stream logging failed: %v", err)
	} else {
		log.Printf("[WAL] Atomic WAL entry pushed to Redis Stream for transfer %s", transferID)
	}

	// 4. Create the TransferTransaction with PENDING status in GORM
	txn := &ledger.TransferTransaction{
		SourceAccountID:      req.SourceAccountId,
		DestinationAccountID: req.DestinationAccountId,
		Amount:               req.Amount,
		Currency:             req.Currency,
		Status:               "PENDING",
	}
	// Autofill ExtID
	txn.Autofill(txn)
	// Override ExtID to match the transferID if provided
	if req.TransferId != "" {
		txn.ExtID = &req.TransferId
	}
	
	if err := s.app.DB.Create(txn).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create transfer transaction: %v", err)
	}

	// Return response with PENDING status
	return &ledgerpb.TransferResponse{
		TransferId:  *txn.ExtID,
		Status:      "PENDING",
		ErrorReason: "",
	}, nil
}

func (s *LedgerServer) UpdateTransferStatus(ctx context.Context, req *ledgerpb.UpdateStatusRequest) (*ledgerpb.UpdateStatusResponse, error) {
	var txn ledger.TransferTransaction
	err := s.app.DB.Where("ext_id = ?", req.TransferId).First(&txn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "transfer not found")
		}
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	oldStatus := txn.Status
	newStatus := req.NewStatus

	// Enforce sequential state transitions: PENDING -> APPROVED -> PROCESSING -> COMPLETED
	valid := false
	switch oldStatus {
	case "PENDING":
		valid = (newStatus == "APPROVED")
	case "APPROVED":
		valid = (newStatus == "PROCESSING")
	case "PROCESSING":
		valid = (newStatus == "COMPLETED")
	}

	if !valid {
		log.Printf("[State Machine] Invalid transition from %s to %s attempted for transfer %s", oldStatus, newStatus, req.TransferId)
		return nil, status.Errorf(codes.FailedPrecondition, "INVALID_TRANSFER_STATE")
	}

	// If transitioning to COMPLETED, execute GORM optimistic lock transfer logic
	if newStatus == "COMPLETED" {
		maxRetries := 5
		var errTx error

		for i := 0; i < maxRetries; i++ {
			errTx = s.app.DB.Transaction(func(tx *gorm.DB) error {
				var src ledger.Account
				var dest ledger.Account

				// Fetch source
				if err := tx.Where("ext_id = ?", txn.SourceAccountID).First(&src).Error; err != nil {
					return status.Errorf(codes.NotFound, "source account not found")
				}

				// Fetch destination
				if err := tx.Where("ext_id = ?", txn.DestinationAccountID).First(&dest).Error; err != nil {
					return status.Errorf(codes.NotFound, "destination account not found")
				}

				// Validate sufficient funds
				if src.Balance < txn.Amount {
					return status.Errorf(codes.FailedPrecondition, "insufficient funds")
				}

				// Version details for optimistic lock
				oldSrcVersion := src.Version
				oldDestVersion := dest.Version

				// Deduct from source
				src.Balance -= txn.Amount
				src.Version += 1
				resSrc := tx.Model(&ledger.Account{}).
					Where("ext_id = ? AND version = ?", txn.SourceAccountID, oldSrcVersion).
					Updates(map[string]interface{}{
						"balance": src.Balance,
						"version": src.Version,
					})

				if resSrc.Error != nil {
					return resSrc.Error
				}
				if resSrc.RowsAffected == 0 {
					return gorm.ErrInvalidDB // Version conflict
				}

				// Add to destination
				dest.Balance += txn.Amount
				dest.Version += 1
				resDest := tx.Model(&ledger.Account{}).
					Where("ext_id = ? AND version = ?", txn.DestinationAccountID, oldDestVersion).
					Updates(map[string]interface{}{
						"balance": dest.Balance,
						"version": dest.Version,
					})

				if resDest.Error != nil {
					return resDest.Error
				}
				if resDest.RowsAffected == 0 {
					return gorm.ErrInvalidDB // Version conflict
				}

				return nil
			})

			if errTx == nil {
				// Invalidate Redis Cache
				s.app.Cache.Del(ctx, fmt.Sprintf("balance:%s", txn.SourceAccountID))
				s.app.Cache.Del(ctx, fmt.Sprintf("balance:%s", txn.DestinationAccountID))
				break
			}

			// If it's a version conflict error, run backoff retry
			if errors.Is(errTx, gorm.ErrInvalidDB) {
				backoff := time.Duration(math.Pow(2, float64(i))) * 50 * time.Millisecond
				time.Sleep(backoff)
				continue
			}

			// For other errors (insufficient funds, etc.), abort
			break
		}

		if errTx != nil {
			// Failed to execute transfer
			log.Printf("[State Machine] Transfer execution failed for %s: %v", req.TransferId, errTx)
			
			// Emit transaction.failed event to RabbitMQ
			failedPayload := map[string]interface{}{
				"transfer_id":  req.TransferId,
				"error_reason": errTx.Error(),
			}
			s.app.MQ.Emit("transaction.failed", failedPayload)

			return nil, errTx
		}

		// Emit transaction.completed event to RabbitMQ
		eventPayload := map[string]interface{}{
			"transfer_id": req.TransferId,
			"amount":      txn.Amount,
			"status":      "COMPLETED",
		}
		s.app.MQ.Emit("transaction.completed", eventPayload)
	}

	// Update Status in DB
	txn.Status = newStatus
	if err := s.app.DB.Save(&txn).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update transfer transaction: %v", err)
	}

	return &ledgerpb.UpdateStatusResponse{
		TransferId: req.TransferId,
		Status:     newStatus,
	}, nil
}

func (s *LedgerServer) BatchTransfer(ctx context.Context, req *ledgerpb.BatchTransferRequest) (*ledgerpb.BatchTransferResponse, error) {
	// Validation: Max 10,000 transfers
	if len(req.Transfers) > 10000 {
		return nil, status.Errorf(codes.InvalidArgument, "bulk distributions capped at 10,000 transfers maximum")
	}

	var responses []*ledgerpb.TransferResponse
	for _, tReq := range req.Transfers {
		res, err := s.Transfer(ctx, tReq)
		if err != nil {
			responses = append(responses, &ledgerpb.TransferResponse{
				TransferId:  tReq.TransferId,
				Status:      "FAILED",
				ErrorReason: err.Error(),
			})
		} else {
			responses = append(responses, res)
		}
	}

	return &ledgerpb.BatchTransferResponse{
		Responses: responses,
	}, nil
}
