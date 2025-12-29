package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email      string             `bson:"email" json:"email"`
	Name       string             `bson:"name" json:"name"`
	PictureURL string             `bson:"picture_url" json:"picture_url"`
	Role       Role               `bson:"role" json:"role"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type Movie struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Genre       string             `bson:"genre" json:"genre"`
	DurationMin int                `bson:"duration_min" json:"duration_min"`
	PosterURL   string             `bson:"poster_url" json:"poster_url"`
	Screenings  []Screening        `bson:"screenings" json:"screenings"`
}

type Screening struct {
	ID        string    `bson:"id" json:"id"` // Unique ID generator needed or use UUID
	StartTime time.Time `bson:"start_time" json:"start_time"`
	Price     float64   `bson:"price" json:"price"`
	Seats     []Seat    `bson:"seats" json:"seats,omitempty"`
}

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatBooked    SeatStatus = "BOOKED"
)

type Seat struct {
	ID       string     `bson:"id" json:"id"`
	Row      string     `bson:"row" json:"row"`
	Number   int        `bson:"number" json:"number"`
	Status   SeatStatus `bson:"status" json:"status"`
	LockedBy string     `bson:"locked_by,omitempty" json:"locked_by,omitempty"`
}

type Booking struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	ScreeningID     string             `bson:"screening_id" json:"screening_id"`
	ScreenStartTime string             `bson:"screen_start_time" json:"screen_start_time"`
	SeatID          string             `bson:"seat_id" json:"seat_id"`
	Status          string             `bson:"status" json:"status"`
	PaymentID       string             `bson:"payment_id" json:"payment_id"` // [NEW] Payment Reference
	Amount          float64            `bson:"amount" json:"amount"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}
