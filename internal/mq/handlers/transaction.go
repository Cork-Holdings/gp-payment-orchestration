package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/rabbitmq/amqp091-go"
)

var (
	ErrFatalCrossWallet = errors.New("CROSS_WALLET_NOT_ALLOWED")
	ErrFatalIdempotency = errors.New("IDEMPOTENCY_CONFLICT")
)

func isFatal(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrFatalCrossWallet) || 
		errors.Is(err, ErrFatalIdempotency) || 
		strings.Contains(err.Error(), "CROSS_WALLET_NOT_ALLOWED") || 
		strings.Contains(err.Error(), "IDEMPOTENCY_CONFLICT")
}

func HandleTransactionCompleted(app *global.App, msg amqp091.Delivery) error {
	maxRetries := 5
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = processTransactionCompleted(app, msg.Body)
		if err == nil {
			log.Printf("[MQ] Successfully processed transaction.completed event on attempt %d", attempt)
			return nil
		}

		if isFatal(err) {
			log.Printf("[MQ] Fatal error encountered in transaction.completed: %v. Rejecting to DLQ immediately.", err)
			return err
		}

		if attempt < maxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
			log.Printf("[MQ] Transient error in transaction.completed (attempt %d/%d): %v. Retrying in %v...", attempt, maxRetries, err, backoff)
			time.Sleep(backoff)
		}
	}

	log.Printf("[MQ] Retries exhausted for transaction.completed. Error: %v. Directing to DLQ.", err)
	return err
}

func HandleTransactionFailed(app *global.App, msg amqp091.Delivery) error {
	maxRetries := 3
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = processTransactionFailed(app, msg.Body)
		if err == nil {
			log.Printf("[MQ] Successfully processed transaction.failed event on attempt %d", attempt)
			return nil
		}

		if isFatal(err) {
			log.Printf("[MQ] Fatal error in transaction.failed: %v. Rejecting to DLQ.", err)
			return err
		}

		if attempt < maxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
			time.Sleep(backoff)
		}
	}

	return err
}

func HandleSettlementCompleted(app *global.App, msg amqp091.Delivery) error {
	maxRetries := 3
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = processSettlementCompleted(app, msg.Body)
		if err == nil {
			log.Printf("[MQ] Successfully processed settlement.completed event on attempt %d", attempt)
			return nil
		}

		if isFatal(err) {
			return err
		}

		if attempt < maxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
			time.Sleep(backoff)
		}
	}

	return err
}

func HandleKycApproved(app *global.App, msg amqp091.Delivery) error {
	err := processKycApproved(app, msg.Body)
	if err != nil {
		log.Printf("[MQ] KYC approved event failed: %v. No retries allowed. Routing to DLQ.", err)
		return err
	}
	log.Printf("[MQ] Successfully processed kyc.approved event")
	return nil
}

func processTransactionCompleted(app *global.App, body []byte) error {
	var payload struct {
		TransferID string  `json:"transfer_id"`
		Amount     float64 `json:"amount"`
		Status     string  `json:"status"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if payload.TransferID == "fatal_cross_wallet" {
		return ErrFatalCrossWallet
	}
	if payload.TransferID == "fatal_idempotency" {
		return ErrFatalIdempotency
	}
	if payload.TransferID == "transient_error" && time.Now().UnixNano()%2 == 0 {
		return fmt.Errorf("transient database connection timeout")
	}

	log.Printf("[TimescaleDB Simulation] Logging completed transfer %s, amount %f to analytics database", payload.TransferID, payload.Amount)
	log.Printf("[Webhook Gateway Simulation] Triggering external callback to merchant for transfer %s", payload.TransferID)
	
	common.LogAuditEvent(app, "webhook_gateway", "transaction.completed", payload)
	return nil
}

func processTransactionFailed(app *global.App, body []byte) error {
	var payload struct {
		TransferID  string `json:"transfer_id"`
		ErrorReason string `json:"error_reason"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	log.Printf("[Reporting Simulation] Logged transfer failure for %s. Reason: %s", payload.TransferID, payload.ErrorReason)
	
	common.LogAuditEvent(app, "reporting_service", "transaction.failed", payload)
	return nil
}

func processSettlementCompleted(app *global.App, body []byte) error {
	var payload struct {
		SettlementID string  `json:"settlement_id"`
		NetAmount    float64 `json:"net_amount"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	log.Printf("[Settlement Simulation] Settlement %s completed for net amount %f", payload.SettlementID, payload.NetAmount)
	
	common.LogAuditEvent(app, "settlement_engine", "settlement.completed", payload)
	return nil
}

func processKycApproved(app *global.App, body []byte) error {
	var payload struct {
		TenantID      string `json:"tenant_id"`
		ApprovalLevel string `json:"approval_level"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	log.Printf("[Compliance Simulation] KYC approved for tenant %s at level %s", payload.TenantID, payload.ApprovalLevel)
	
	common.LogAuditEvent(app, "kyc_compliance_service", "kyc.approved", payload)
	return nil
}
