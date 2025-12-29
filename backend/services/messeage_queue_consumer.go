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

// Consumer คือ Struct ที่ทำหน้าที่เป็นตัวแทนของ Consumer Group (คนรับสารแบบกลุ่ม)
type Consumer struct{}

// Setup ทำงาน 1 ครั้งตอนเริ่ม Session (ก่อนเริ่มดึงข้อความ)
// ส่วนใหญ่เอาไว้ Connect Database หรือเตรียมตัวแปร
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup ทำงาน 1 ครั้งตอนจบ Session (หลังจากหยุดดึงข้อความ)
// เอาไว้ปิด Connection หรือ Save State ครั้งสุดท้าย
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim คือ "หัวใจหลัก" ของการทำงาน
// ฟังก์ชันนี้จะถูกเรียกเมื่อ Consumer ได้รับสิทธิ์ในการอ่าน Partition นั้นๆ
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE: ห้ามใช้ go routine ในนี้ (เช่น go func()...) เพราะ library มันจัดการให้แล้ว
	// เราแค่ loop อ่านข้อความจาก claim.Messages() ก็พอ
	for message := range claim.Messages() {
		// 1. ประมวลผลข้อความ (เช่น ส่งเมล, บันทึก logs)
		handleMQMessage(message)

		// 2. บอก Kafka ว่า "ทำเสร็จแล้วนะ" (Mark Offset)
		// Kafka จะจดไว้ว่ากลุ่มนี้อ่านถึงไหนแล้ว ถ้า Restart จะได้มาทำต่อจากตรงนี้ ไม่เริ่มใหม่แต่ต้น
		session.MarkMessage(message, "")
	}
	return nil
}

// StartQueueConsumer เริ่มต้นระบบรับข้อความแบบ Consumer Group (ทำงานเบื้องหลัง)
func StartQueueConsumer() {
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true
	// OffsetOldest: ถ้าเป็นกลุ่มใหม่ที่ไม่เคยจดจำ ให้เริ่มอ่านตั้งแต่ข้อความแรกสุด (กันตกหล่น)
	// แต่ถ้าเคยจดจำแล้ว (MarkMessage) มันจะอ่านต่อจากเดิมให้อัตโนมัติ
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	var consumerGroup sarama.ConsumerGroup
	var err error

	// วนลูปพยายามเชื่อมต่อ Kafka (เผื่อ Kafka ยังไม่ตื่น)
	for i := 0; i < 10; i++ {
		// GROUP ID: "movie-ticket-backend-group"
		// สำคัญมาก! ต้องตั้งชื่อกลุ่มให้เหมือนเดิมตลอด Kafka ถึงจะจำได้ว่ากลุ่มนี้อ่านถึงไหนแล้ว
		consumerGroup, err = sarama.NewConsumerGroup([]string{config.AppConfig.KafkaBrokers}, "movie-ticket-backend-group", cfg)
		if err == nil {
			break
		}
		log.Printf("Failed to start Kafka consumer group, retrying... %v", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Printf("FINAL: Failed to connect Kafka consumer. MQ consumer disabled.")
		return
	}

	fmt.Println("MQ: Queue Consumer Group started (Kafka)...")

	// รันลูปการทำงานใน Background (Goroutine)
	ctx := context.Background()
	consumer := &Consumer{}

	go func() {
		defer consumerGroup.Close()
		for {
			// userGroup.Consume เป็น Blocking Call (ทำงานค้างไว้ยาวๆ)
			// ถ้ามีการ Rebalance หรือ Connection หลุด มันจะ return error ออกมา
			if err := consumerGroup.Consume(ctx, []string{"booking_events"}, consumer); err != nil {
				log.Printf("MQ Error from consumer: %v", err)
				time.Sleep(2 * time.Second) // รอแป๊บแล้วค่อย connect ใหม่
			}
			// ถ้า Context ถูกยกเลิก (เช่น ปิดโปรแกรม) ให้จบการทำงาน
			if ctx.Err() != nil {
				return
			}
		}
	}()
}

// handleMQMessage ฟังก์ชันแยกประเภทข้อความและส่งไปทำงานต่อ
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
		// ถ้าเป็นประวัติระบบ -> บันทึกลง MongoDB
		saveAuditToMongo(event.Payload)
	default:
		log.Printf("MQ [IGNORED]: Unknown event type: %s", event.Type)
	}
}

// triggerNotification จำลองการส่งแจ้งเตือน (Email/SMS)
func triggerNotification(payload interface{}) {
	// ของจริงอาจจะยิง API ไปหา SendGrid / Twilio
	log.Printf("MQ [NOTIFICATION]: Sending confirmation email... Booking Details: %v", payload)
}

// saveAuditToMongo บันทึกข้อมูลลง Audit Log ใน MongoDB
func saveAuditToMongo(payload interface{}) {
	if database.Mongo == nil {
		log.Printf("MQ [LOG ERROR]: MongoDB connection is NIL")
		return
	}
	collection := database.Mongo.Collection("audit_logs")

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

	// ถ้าไม่มี ID ให้สร้างใหม่ (กัน Error)
	if logEntry.ID.IsZero() {
		logEntry.ID = primitive.NewObjectID()
	}

	log.Printf("MQ [DEBUG]: Attempting to insert AuditLog: %+v", logEntry)

	_, err = collection.InsertOne(context.Background(), logEntry)
	if err != nil {
		log.Printf("MQ [LOG ERROR]: Failed to save to Mongo: %v", err)
	} else {
		log.Printf("MQ [LOG SUCCESS]: Audit log saved to Mongo (Async via Kafka)")
	}
}
