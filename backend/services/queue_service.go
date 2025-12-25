package services

import (
	"encoding/json"
	"log"
	"movie-ticket-backend/config"
	"time"

	"github.com/IBM/sarama"
)

type QueueService struct {
	Producer sarama.SyncProducer
}

var queueService *QueueService

func InitQueueService() {
	// Kafka Config
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5

	var producer sarama.SyncProducer
	var err error

	// Retry loop for Docker startup
	for i := 0; i < 10; i++ {
		producer, err = sarama.NewSyncProducer([]string{config.AppConfig.KafkaBrokers}, cfg)
		if err == nil {
			break
		}
		log.Printf("Failed to start Kafka producer, retrying... %v", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Printf("FINAL: Failed to connect to Kafka: %v", err)
		return
	}

	queueService = &QueueService{
		Producer: producer,
	}
	log.Println("Kafka Producer Connected")
}

func GetQueueService() *QueueService {
	return queueService
}

type BookingEvent struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	CreatedAt time.Time   `json:"created_at"`
}

func (s *QueueService) PublishEvent(eventType string, payload interface{}) error {
	if s.Producer == nil {
		return nil
	}

	event := BookingEvent{
		Type:      eventType,
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	val, _ := json.Marshal(event)

	msg := &sarama.ProducerMessage{
		Topic: "booking_events",
		Value: sarama.StringEncoder(val),
	}

	partition, offset, err := s.Producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return err
	}
	log.Printf("Message sent to partition %d at offset %d", partition, offset)
	return nil
}
