package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"movie-ticket-backend/config"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"time"

	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StartQueueConsumer ทำหน้าที่เปิด "หู" คอยฟังข่าวสารจาก Kafka แบบทำงานเบื้องหลัง (Background)
func StartQueueConsumer() {
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest // อ่านตั้งแต่ข้อความแรกสุด (Backlog)

	var consumer sarama.Consumer
	var err error

	// วนลูปพยายามเริ่มระบบ (เช่นเดียวกับฝั่ง Producer)
	for i := 0; i < 10; i++ {
		consumer, err = sarama.NewConsumer([]string{config.AppConfig.KafkaBrokers}, cfg)
		if err == nil {
			break
		}
		log.Printf("Failed to start Kafka consumer, retrying... %v", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Printf("FINAL: Failed to connect Kafka consumer. MQ consumer disabled.")
		return
	}

	// เลือกฟังเฉพาะ Topic "booking_events" (Partition 0)
	partitionConsumer, err := consumer.ConsumePartition("booking_events", 0, sarama.OffsetOldest)
	if err != nil {
		log.Printf("Failed to consume partition: %v", err)
		return
	}

	fmt.Println("MQ: Queue Consumer started (Kafka)...")

	// รันลูปเช็คข้อความเข้าแบบตลอดเวลา
	go func() {
		defer consumer.Close()
		defer partitionConsumer.Close()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				// เมื่อมีจดหมายเข้ามา ส่งไปประมวลผลต่อ
				handleMQMessage(msg)
			case err := <-partitionConsumer.Errors():
				log.Printf("MQ Error: %v", err)
			}
		}
	}()
}

// handleMQMessage ทำหน้าที่ "คัดแยกประเภท" ของจดหมายที่ได้รับ
func handleMQMessage(msg *sarama.ConsumerMessage) {
	log.Printf("MQ [RAW]: Received message value: %s", string(msg.Value))

	var event BookingEvent
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		log.Printf("Failed to unmarshal MQ event: %v", err)
		return
	}

	fmt.Printf("MQ [RECEIVED]: Type=%s\n", event.Type)

	switch event.Type {
	case "BOOKING_SUCCESS":
		// ถ้าจองสำเร็จ -> ส่งแจ้งเตือนลูกค้า
		triggerNotification(event.Payload)
	case "AUDIT_LOG":
		// ถ้าเป็นประวัติระบบ -> บันทึกลง MongoDB แบบ Async
		saveAuditToMongo(event.Payload)
	default:
		log.Printf("MQ [IGNORED]: Unknown event type: %s", event.Type)
	}
}

// triggerNotification (Requirement) จำลองการส่งข้อความแจ้งเตือน (Email/SMS)
func triggerNotification(payload interface{}) {
	// ในระบบจริง จะเรียกใช้ API ของพวก SendGrid หรือ Twilio ที่นี่
	log.Printf("MQ [NOTIFICATION]: Sending confirmation email... Booking Details: %v", payload)
}

// saveAuditToMongo (Requirement) บันทึกประวัติลง MongoDB แบบทำงานเบื้องหลัง (Async Logging)
func saveAuditToMongo(payload interface{}) {
	if database.Mongo == nil {
		log.Printf("MQ [LOG ERROR]: MongoDB connection is NIL")
		return
	}
	collection := database.Mongo.Collection("audit_logs")

	// แปลงข้อมูล Payload กลับเป็นรูปแบบ AuditLog
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("MQ [LOG ERROR]: Failed to marshal payload: %v", err)
		return
	}

	var logEntry models.AuditLog
	err = json.Unmarshal(data, &logEntry)
	if err != nil {
		log.Printf("MQ [LOG ERROR]: Failed to unmarshal to AuditLog: %v", err)
		return
	}

	// FIX: Handle ObjectID explicitly if it's zero
	if logEntry.ID.IsZero() {
		logEntry.ID = primitive.NewObjectID()
	}

	log.Printf("MQ [DEBUG]: Attempting to insert AuditLog: %+v", logEntry)

	// บันทึกลงฐานข้อมูล
	_, err = collection.InsertOne(context.Background(), logEntry)
	if err != nil {
		log.Printf("MQ [LOG ERROR]: Failed to save to Mongo: %v", err)
	} else {
		log.Printf("MQ [LOG SUCCESS]: Audit log saved to Mongo (Async via Kafka)")
	}
}
