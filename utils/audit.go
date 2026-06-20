package utils

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
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

var (
	jsonKeyRegex = regexp.MustCompile(`(?i)"(card_number|card|phone_number|phone|mobile|nrc|national_id)"\s*:\s*"([^"]+)"`)
	cardRegex    = regexp.MustCompile(`\b\d{13,19}\b`)
	nrcRegex     = regexp.MustCompile(`\b\d{6}/\d{2}/\d\b`)
	phoneRegex   = regexp.MustCompile(`\b(?:\+?260|0)[79]\d{8}\b`)
)

func maskPII(payload string) string {
	payload = jsonKeyRegex.ReplaceAllStringFunc(payload, func(match string) string {
		submatches := jsonKeyRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		key := submatches[1]
		val := submatches[2]

		maskedVal := "[MASKED]"
		lowerKey := strings.ToLower(key)
		if strings.Contains(lowerKey, "card") {
			if len(val) > 10 {
				maskedVal = val[:6] + strings.Repeat("*", len(val)-10) + val[len(val)-4:]
			} else {
				maskedVal = "[MASKED_CARD]"
			}
		} else if strings.Contains(lowerKey, "nrc") {
			maskedVal = "[MASKED_NRC]"
		} else if strings.Contains(lowerKey, "phone") || strings.Contains(lowerKey, "mobile") {
			if len(val) > 4 {
				maskedVal = val[:3] + strings.Repeat("*", len(val)-4) + val[len(val)-1:]
			} else {
				maskedVal = "[MASKED_MOBILE]"
			}
		}
		return `"` + key + `":"` + maskedVal + `"`
	})

	payload = cardRegex.ReplaceAllString(payload, "[MASKED_CARD]")
	payload = nrcRegex.ReplaceAllString(payload, "[MASKED_NRC]")
	payload = phoneRegex.ReplaceAllString(payload, "[MASKED_MOBILE]")

	return payload
}

func LogAuditEvent(app *global.App, serviceName string, eventName string, payload interface{}) {
	payloadBytes, _ := json.Marshal(payload)
	maskedPayload := maskPII(string(payloadBytes))

	logEntry := AuditLog{
		Event:     eventName,
		Service:   serviceName,
		Payload:   maskedPayload,
		Timestamp: time.Now(),
	}

	_, err := repo.MongoCreateOne(app, logEntry)
	if err != nil {
		log.Printf("[AuditLog] Failed to write audit log to MongoDB: %v", err)
	} else {
		log.Printf("[AuditLog] Audit log successfully written to MongoDB for event: %s", eventName)
	}
}
