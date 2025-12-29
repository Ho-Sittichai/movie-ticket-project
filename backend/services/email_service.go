package services

import (
	"fmt"
	"log"
	"movie-ticket-backend/models"
	"net/smtp"
	"os"
	"strings"

	"movie-ticket-backend/config"
)

// EmailService provides methods to send emails
type EmailService struct{}

var emailService = &EmailService{}

func GetEmailService() *EmailService {
	return emailService
}

// SendGroupTicketEmail simulates sending a single email for multiple tickets
func (s *EmailService) SendGroupTicketEmail(user models.User, bookings []models.Booking, movieTitle string) {
	if len(bookings) == 0 {
		return
	}

	firstBooking := bookings[0]
	totalAmount := 0.0
	var seatList []string

	// Calculate totals and collect seats
	for _, b := range bookings {
		totalAmount += b.Amount
		seatList = append(seatList, b.SeatID)
	}
	seatsStr := strings.Join(seatList, ", ")

	// Email Content
	subject := fmt.Sprintf("Your Tickets for %s", movieTitle)

	body := new(strings.Builder)
	body.WriteString(fmt.Sprintf("To: %s\r\n", user.Email))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	body.WriteString("\r\n") // End of headers

	body.WriteString("==================================================\n")
	body.WriteString(fmt.Sprintf("             MOVIE TICKETS CONFIRMED          \n"))
	body.WriteString("==================================================\n")
	body.WriteString(fmt.Sprintf(" Hello %s,\n", user.Name))
	body.WriteString("\n")
	body.WriteString(" Thank you for your purchase! Here are your ticket details:\n")
	body.WriteString("\n")
	body.WriteString(fmt.Sprintf(" Movie:      %s\n", movieTitle))
	body.WriteString(fmt.Sprintf(" Show Time:  %s\n", firstBooking.ScreenStartTime))
	body.WriteString(fmt.Sprintf(" Seats:      %s\n", seatsStr))
	body.WriteString(fmt.Sprintf(" Total Price: %.2f THB\n", totalAmount))
	body.WriteString(fmt.Sprintf(" Order IDs:  %s ...\n", firstBooking.ID.Hex()))
	body.WriteString("\n")
	body.WriteString("--------------------------------------------------\n")
	body.WriteString(" Please show this email at the theater entrance.\n")
	body.WriteString("==================================================\n")

	// Check if real email config is available
	if config.AppConfig.GoogleClientID != "" || os.Getenv("EMAIL_SENDER") != "" {
		// Ideally checking specific EMAIL_SENDER config from env directly as it wasn't in config struct yet
		// But for now, we follow the "Mock/Simulate" request from user but prepare logic if they add env vars
		sender := os.Getenv("EMAIL_SENDER")
		password := os.Getenv("EMAIL_PASSWORD")

		if sender != "" && password != "" {
			auth := smtp.PlainAuth("", sender, password, "smtp.gmail.com")
			to := []string{user.Email}
			msg := []byte(body.String())

			err := smtp.SendMail("smtp.gmail.com:587", auth, sender, to, msg)
			if err != nil {
				log.Printf("[EMAIL ERROR] Failed to send real email: %v", err)
			} else {
				log.Printf("[EMAIL SENT] Real email sent to %s", user.Email)
				return // Exit if sent successfully
			}
		}
	}

	// Fallback to Console Log (Mock)
	log.Printf("\n[EMAIL SENT] To: %s (%s)\nSubject: %s%s", user.Name, user.Email, subject, body.String())
}
