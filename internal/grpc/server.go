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

	maxRetries := 5
	var errTx error

	for i := 0; i < maxRetries; i++ {
		errTx = s.app.DB.Transaction(func(tx *gorm.DB) error {
			var src ledger.Account
			var dest ledger.Account

			// Fetch source
			if err := tx.Where("ext_id = ?", req.SourceAccountId).First(&src).Error; err != nil {
				return status.Errorf(codes.NotFound, "source account not found")
			}

			// Fetch destination
			if err := tx.Where("ext_id = ?", req.DestinationAccountId).First(&dest).Error; err != nil {
				return status.Errorf(codes.NotFound, "destination account not found")
			}

			// Step 3: Inject Wallet-Type validation check (return CROSS_WALLET_NOT_ALLOWED)
			if src.WalletType != dest.WalletType {
				return status.Errorf(codes.InvalidArgument, "CROSS_WALLET_NOT_ALLOWED")
			}

			// Validate sufficient funds
			if src.Balance < req.Amount {
				return status.Errorf(codes.FailedPrecondition, "insufficient funds")
			}

			// Version details for optimistic lock
			oldSrcVersion := src.Version
			oldDestVersion := dest.Version

			// Deduct from source
			src.Balance -= req.Amount
			src.Version += 1
			resSrc := tx.Model(&ledger.Account{}).
				Where("ext_id = ? AND version = ?", req.SourceAccountId, oldSrcVersion).
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
			dest.Balance += req.Amount
			dest.Version += 1
			resDest := tx.Model(&ledger.Account{}).
				Where("ext_id = ? AND version = ?", req.DestinationAccountId, oldDestVersion).
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
			// Success! Invalidate Redis cache
			s.app.Cache.Del(ctx, fmt.Sprintf("balance:%s", req.SourceAccountId))
			s.app.Cache.Del(ctx, fmt.Sprintf("balance:%s", req.DestinationAccountId))

			// Emit transaction.completed event to RabbitMQ
			eventPayload := map[string]interface{}{
				"transfer_id": transferID,
				"amount":      req.Amount,
				"status":      "COMPLETED",
			}
			if emitErr := s.app.MQ.Emit("transaction.completed", eventPayload); emitErr != nil {
				log.Printf("Failed to publish transaction.completed event: %v", emitErr)
			}

			return &ledgerpb.TransferResponse{
				TransferId:  transferID,
				Status:      "COMPLETED",
				ErrorReason: "",
			}, nil
		}

		// If it's a version conflict error, run exponential backoff retry
		if errors.Is(errTx, gorm.ErrInvalidDB) {
			backoff := time.Duration(math.Pow(2, float64(i))) * 50 * time.Millisecond
			log.Printf("Version conflict on transfer. Retrying in %v...", backoff)
			time.Sleep(backoff)
			continue
		}

		// Insufficient funds, wallet validation check failure or account not found
		break
	}

	// If failed due to insufficient funds, wallet check, or retries exhausted
	var statusMsg string
	var errReason string
	var grpcErr error

	if errors.Is(errTx, gorm.ErrInvalidDB) {
		statusMsg = "FAILED"
		errReason = "CONCURRENT_MODIFICATION_ERROR"
		grpcErr = status.Errorf(codes.Aborted, "concurrent modification conflict")
	} else {
		statusMsg = "FAILED"
		errReason = errTx.Error()
		grpcErr = errTx
	}

	// Emit transaction.failed event to RabbitMQ
	failedPayload := map[string]interface{}{
		"transfer_id":  transferID,
		"error_reason": errReason,
	}
	if emitErr := s.app.MQ.Emit("transaction.failed", failedPayload); emitErr != nil {
		log.Printf("Failed to publish transaction.failed event: %v", emitErr)
	}

	if errors.Is(errTx, gorm.ErrInvalidDB) {
		return nil, grpcErr
	}

	return &ledgerpb.TransferResponse{
		TransferId:  transferID,
		Status:      statusMsg,
		ErrorReason: errReason,
	}, grpcErr
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
