package services

import (
	"encoding/json"
	"log"
	"movie-ticket-backend/config"
	"time"

	"github.com/IBM/sarama"
)

// QueueService คือตัวแทนของระบบส่งข้อความ (Producer)
type QueueService struct {
	Producer sarama.SyncProducer
}

var queueService *QueueService

// InitQueueService ทำหน้าที่ตั้งค่าและเชื่อมต่อกับ Kafka (ฝ่ายส่ง)
func InitQueueService() {
	// สร้าง Instance ทันทีเพื่อให้ GetQueueService() ไม่เป็น nil แม้จะต่อ Kafka ไม่ติด
	queueService = &QueueService{}

	// ตั้งค่าคอนฟิกของ Kafka
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5

	var producer sarama.SyncProducer
	var err error

	// วนลูปพยายามเริ่มระบบ (เผื่อ Docker รัน Kafka ช้ากว่าโค้ด)
	for i := 0; i < 10; i++ {
		producer, err = sarama.NewSyncProducer([]string{config.AppConfig.KafkaBrokers}, cfg)
		if err == nil {
			queueService.Producer = producer
			log.Println("Kafka Producer Connected")
			return
		}
		log.Printf("Failed to start Kafka producer, retrying... %v", err)
		time.Sleep(2 * time.Second)
	}

	log.Printf("FINAL: Failed to connect to Kafka. MQ features will be disabled.")
}

// GetQueueService ใช้สำหรับดึง Instance ของตัวส่งไปใช้งานในไฟล์อื่น
func GetQueueService() *QueueService {
	if queueService == nil {
		return &QueueService{} // ส่งกลับก้อนว่างๆ เพื่อป้องกัน Panic
	}
	return queueService
}

// BookingEvent คือโครงสร้างของข้อมูลที่จะส่งเข้า Kafka
type BookingEvent struct {
	Type      string      `json:"type"`       // ประเภท (เช่น BOOKING_SUCCESS, AUDIT_LOG)
	Payload   interface{} `json:"payload"`    // ข้อมูลไส้ใน
	CreatedAt time.Time   `json:"created_at"` // เวลาที่เกิดเหตุการณ์
}

// PublishEvent คือฟังก์ชัน "ส่งจดหมาย" โดยจะเอาข้อมูลโยนใส่ Topic "booking_events"
func (s *QueueService) PublishEvent(eventType string, payload interface{}) error {
	if s == nil || s.Producer == nil {
		return nil // ข้ามการส่งถ้า Kafka ไม่พร้อม (ป้องกัน Panic)
	}

	event := BookingEvent{
		Type:      eventType,
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	// แปลงข้อมูลเป็น JSON
	val, _ := json.Marshal(event)

	msg := &sarama.ProducerMessage{
		Topic: "booking_events", // หัวข้อที่จะส่งไป
		Value: sarama.StringEncoder(val),
	}

	// ส่งข้อมูลเข้า Kafka
	partition, offset, err := s.Producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return err
	}
	log.Printf("Message sent to partition %d at offset %d", partition, offset)
	return nil
}
