package common

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/repo"
)

type AuditLog struct {
	Event     string    `bson:"event" json:"event"`
	Service   string    `bson:"service" json:"service"`
	Payload   string    `bson:"payload" json:"payload"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

func (a AuditLog) CollectionName() string {
	return "audit_logs"
}

func LogAuditEvent(app *global.App, serviceName string, eventName string, payload interface{}) {
	payloadBytes, _ := json.Marshal(payload)
	
	logEntry := AuditLog{
		Event:     eventName,
		Service:   serviceName,
		Payload:   string(payloadBytes),
		Timestamp: time.Now(),
	}
	
	_, err := repo.MongoCreateOne(app, logEntry)
	if err != nil {
		log.Printf("[AuditLog] Failed to write audit log to MongoDB: %v", err)
	} else {
		log.Printf("[AuditLog] Audit log successfully written to MongoDB for event: %s", eventName)
	}
}
