package services

import (
	"context"
	"fmt"
	"log"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"time"
)

// Global Channel for buffering logs
var logChannel chan models.AuditLog

const LogBufferSize = 1000

// InitAuditService starts the background worker
func InitAuditService() {
	logChannel = make(chan models.AuditLog, LogBufferSize)

	// Start background worker
	go func() {
		fmt.Println("Audit Log Worker started...")
		collection := database.Mongo.Collection("audit_logs")

		for logEntry := range logChannel {
			_, err := collection.InsertOne(context.Background(), logEntry)
			if err != nil {
				log.Printf("Failed to insert audit log: %v", err)
			}
		}
	}()
}

// LogInfo sends an INFO log to the channel
func LogInfo(eventType string, userID string, details map[string]interface{}) {
	pushLog("INFO", eventType, userID, details)
}

// LogWarn sends a WARN log to the channel
func LogWarn(eventType string, userID string, details map[string]interface{}) {
	pushLog("WARN", eventType, userID, details)
}

// LogError sends an ERROR log to the channel
func LogError(eventType string, userID string, err error, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["error"] = err.Error()
	pushLog("ERROR", eventType, userID, details)
}

func pushLog(level, eventType, userID string, details map[string]interface{}) {
	select {
	case logChannel <- models.AuditLog{
		Timestamp: time.Now(),
		Level:     level,
		EventType: eventType,
		UserID:    userID,
		Details:   details,
	}:
		// Log sent to buffer
	default:
		// Buffer full, drop log to prevent blocking main thread (or handle gracefully)
		log.Printf("Audit Log Buffer Full! Dropping log: %s", eventType)
	}
}
