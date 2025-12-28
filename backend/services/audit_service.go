package services

import (
	"fmt"
	"log"
	"movie-ticket-backend/models"
	"time"
)

// InitAuditService - ไม่จำเป็นต้องมี Background Worker ในเครื่องแล้ว เพราะส่งงานให้ Kafka ทำแทน
func InitAuditService() {
	fmt.Println("Audit Service initialized (Kafka-backed)...")
}

// LogInfo ส่ง Log ประเภทข่าวสารทั่วไปเข้า Kafka
func LogInfo(eventType string, userID string, details map[string]interface{}) {
	pushLog("INFO", eventType, userID, details)
}

// LogWarn ส่ง Log ประเภทคำเตือนเข้า Kafka
func LogWarn(eventType string, userID string, details map[string]interface{}) {
	pushLog("WARN", eventType, userID, details)
}

// LogError ส่ง Log ประเภทข้อผิดพลาดเข้า Kafka
func LogError(eventType string, userID string, err error, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["error"] = err.Error()
	pushLog("ERROR", eventType, userID, details)
}

// pushLog คือไส้ในที่จะสร้างก้อนข้อมูล (AuditLog) แล้วยิงเข้า Kafka Topic "AUDIT_LOG"
func pushLog(level, eventType, userID string, details map[string]interface{}) {
	logEntry := models.AuditLog{
		Timestamp: time.Now(),
		Level:     level,
		EventType: eventType,
		UserID:    userID,
		Details:   details,
	}

	// ใช้ QueueService (Kafka) ในการส่งข้อมูลแบบ Async
	q := GetQueueService()
	if q != nil {
		q.PublishEvent("AUDIT_LOG", logEntry)
	} else {
		// ถ้า Kafka ยังไม่พร้อม (เช่น ช่วงเปิดเครื่อง) ให้พิมพ์ลง Console แก้ขัดไปก่อน
		log.Printf("[AUDIT] %s: %s (User: %s)", level, eventType, userID)
	}
}
