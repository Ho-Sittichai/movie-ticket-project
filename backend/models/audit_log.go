package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Level     string                 `bson:"level" json:"level"`           // INFO, WARN, ERROR
	EventType string                 `bson:"event_type" json:"event_type"` // BOOKING_SUCCESS, SEAT_RELEASED, etc.
	UserID    string                 `bson:"user_id,omitempty" json:"user_id,omitempty"`
	IPAddress string                 `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	Details   map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
}
