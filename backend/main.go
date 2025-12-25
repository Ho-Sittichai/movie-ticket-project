package main

import (
	"context"
	"fmt"
	"log"
	"movie-ticket-backend/config"
	"movie-ticket-backend/database"
	"movie-ticket-backend/handlers"
	"movie-ticket-backend/models"
	"movie-ticket-backend/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	log.Println("Starting Server...")
	config.LoadConfig()

	// Connect DB
	database.ConnectDB()

	// Init Services
	services.InitQueueService() // Kafka
	services.InitWSHub()

	// Seed Data (if needed)
	if database.Mongo != nil {
		SeedData()
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/auth/login", handlers.Login)
		api.GET("/movies", handlers.GetMovies)
		api.POST("/movies", handlers.CreateMovie)
		api.GET("/screenings/:id", handlers.GetScreening)

		api.POST("/seats/lock", handlers.LockSeat)
		api.POST("/seats/book", handlers.BookSeat)
		// api.POST("/seats/unlock", handlers.UnlockSeat) // Implement if needed

		api.GET("/ws", services.ServeWS)
	}

	r.Run(":" + config.AppConfig.Port)
}

func SeedData() {
	moviesColl := database.Mongo.Collection("movies")

	// Define Mock Data to match Frontend
	mockMovies := []struct {
		Title        string
		ScreeningIDs []string
	}{
		{"Avatar: The Way of Water", []string{"s1", "s2", "s3"}},
		{"Oppenheimer", []string{"s4", "s5"}},
		{"Spider-Man: Across the Spider-Verse", []string{"s6", "s7"}},
		{"The Batman", []string{"s8", "s9"}},
		{"Guardians of the Galaxy Vol. 3", []string{"s10"}},
		{"Dune: Part Two", []string{"s11", "s12"}},
		{"Mission: Impossible - Dead Reckoning", []string{"s13", "s14"}},
		{"Barbie", []string{"s15", "s16", "s17"}},
		{"John Wick: Chapter 4", []string{"s18"}},
		{"Inside Out 2", []string{"s19", "s20"}},
	}

	for _, m := range mockMovies {
		var existingMovie models.Movie
		err := moviesColl.FindOne(context.TODO(), bson.M{"title": m.Title}).Decode(&existingMovie)
		if err == nil {
			continue // Already exists
		}

		log.Printf("Seeding %s...", m.Title)

		var screenings []models.Screening
		for _, sid := range m.ScreeningIDs {
			// Create Seats
			var seats []models.Seat
			for r := 0; r < 5; r++ { // 5 Rows
				rowChar := string(rune('A' + r))
				for n := 1; n <= 8; n++ { // 8 Cols
					status := models.SeatAvailable
					// Randomly book
					// if n%3 == 0 { status = models.SeatBooked }
					seats = append(seats, models.Seat{
						ID:     fmt.Sprintf("%s%d", rowChar, n),
						Row:    rowChar,
						Number: n,
						Status: status,
					})
				}
			}

			screenings = append(screenings, models.Screening{
				ID:        sid,
				StartTime: time.Now().Add(time.Hour * 24), // Future
				Price:     200,
				Seats:     seats,
			})
		}

		newMovie := models.Movie{
			ID:          primitive.NewObjectID(),
			Title:       m.Title,
			Description: "Mock Description for " + m.Title,
			DurationMin: 120,
			Screenings:  screenings,
		}
		moviesColl.InsertOne(context.TODO(), newMovie)
	}
	log.Println("All Mock Data Seeded!")
}
