package main

import (
	"context"
	"fmt"
	"log"
	"movie-ticket-backend/config"
	"movie-ticket-backend/database"
	"movie-ticket-backend/handlers"
	"movie-ticket-backend/middleware"
	"movie-ticket-backend/models"
	"movie-ticket-backend/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("Starting Server...")
	config.LoadConfig()

	// Connect DB
	database.ConnectDB()

	// Init Services
	services.InitQueueService()   // Connect Kafka
	services.StartQueueConsumer() // Listen event from kafka
	services.InitWSHub()          // Init WebSocket Hub
	services.InitAuditService()   // Init Audit log Service

	// Start Redis Expiration Listener
	lockService := services.NewLockService()
	go lockService.ListenForExpireRedis()

	// Seed Data (if needed)
	if database.Mongo != nil {
		SeedData()
	}

	r := gin.Default()
	r.Use(middleware.ErrorLogger())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/auth/google/login", handlers.GoogleLogin)
		api.GET("/auth/google/callback", handlers.GoogleCallback)
		api.GET("/movies", handlers.GetMovies)
		api.POST("/movies", handlers.CreateMovie)
		api.POST("/screenings/details", handlers.GetScreeningDetails)

		// Protected Booking Routes
		bookingGroup := api.Group("/seats")
		bookingGroup.Use(middleware.RequireAuth())
		{
			bookingGroup.POST("/lock", handlers.LockSeat)
			bookingGroup.POST("/book", handlers.BookSeat)
			bookingGroup.POST("/extend", handlers.ExtendSeatLock)
		}

		// Protected Payment Routes
		paymentGroup := api.Group("/payment")
		paymentGroup.Use(middleware.RequireAuth())
		{
			paymentGroup.POST("/start", handlers.StartPayment)
			paymentGroup.POST("/cancel", handlers.CancelPayment)
		}
		// api.POST("/seats/unlock", handlers.UnlockSeat) // Implement if needed

		api.GET("/ws", services.ServeWS)
	}

	// Admin API Group (Protected)
	adminAPI := r.Group("/api/admin")
	adminAPI.Use(middleware.AdminAuth())
	{
		adminAPI.GET("/bookings", handlers.GetAllBookings)
	}

	r.Run(":" + config.AppConfig.Port)
}

func getTime(hour, min int) time.Time {
	now := time.Now()
	// Future: Tomorrow at specific hour
	return time.Date(now.Year(), now.Month(), now.Day()+1, hour, min, 0, 0, now.Location())
}

func SeedData() {
	moviesColl := database.Mongo.Collection("movies")

	// Define Mock Data to match Frontend
	mockMovies := []struct {
		Title          string
		Description    string
		Genre          string
		DurationMin    int
		PosterURL      string
		ScreeningTimes []struct {
			ID   string
			Hour int
			Min  int
		}
	}{
		{
			"Avatar: The Way of Water",
			"Jake Sully lives with his newfound family formed on the extrasolar moon Pandora...",
			"Sci-Fi / Action",
			192,
			"https://upload.wikimedia.org/wikipedia/en/5/54/Avatar_The_Way_of_Water_poster.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s1", 10, 0}, {"s2", 14, 0}, {"s3", 18, 0}},
		},
		{
			"Oppenheimer",
			"The story of American scientist J. Robert Oppenheimer...",
			"Biography / Drama",
			180,
			"https://upload.wikimedia.org/wikipedia/en/4/4a/Oppenheimer_%28film%29.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s4", 11, 30}, {"s5", 15, 30}},
		},
		{
			"Spider-Man: Across the Spider-Verse",
			"Miles Morales catapults across the Multiverse...",
			"Animation / Action",
			140,
			"https://upload.wikimedia.org/wikipedia/en/b/b4/Spider-Man-_Across_the_Spider-Verse_poster.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s6", 12, 0}, {"s7", 16, 0}},
		},
		{
			"The Batman",
			"When a sadistic serial killer begins murdering key political figures in Gotham...",
			"Action / Crime",
			176,
			"https://upload.wikimedia.org/wikipedia/en/f/ff/The_Batman_%28film%29_poster.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s8", 19, 0}, {"s9", 22, 0}},
		},
		{
			"Guardians of the Galaxy Vol. 3",
			"Still reeling from the loss of Gamora...",
			"Action / Adventure",
			150,
			"https://upload.wikimedia.org/wikipedia/en/7/74/Guardians_of_the_Galaxy_Vol._3_poster.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s10", 13, 0}},
		},
		{
			"Dune: Part Two",
			"Paul Atreides unites with Chani and the Fremen...",
			"Sci-Fi / Adventure",
			166,
			"https://www.siamzone.com/movie/pic/2024/duneparttwo/poster1.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s11", 14, 30}, {"s12", 18, 30}},
		},
		{
			"Mission: Impossible - Dead Reckoning",
			"Ethan Hunt and his IMF team must track down a dangerous new weapon...",
			"Action / Thriller",
			163,
			"https://theatrgwaun.com/wp-content/uploads/2023/07/TG-Aug23-web-700px-mission-768x768.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s13", 10, 30}, {"s14", 15, 0}},
		},
		{
			"Barbie",
			"Barbie suffers a crisis that leads her to question her world and her existence.",
			"Adventure / Comedy",
			114,
			"https://upload.wikimedia.org/wikipedia/en/0/0b/Barbie_2023_poster.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s15", 11, 0}, {"s16", 13, 30}, {"s17", 16, 0}},
		},
		{
			"John Wick: Chapter 4",
			"John Wick uncovers a path to defeating The High Table.",
			"Action / Crime",
			169,
			"https://assets-prd.ignimgs.com/2023/02/08/jw4-2025x3000-online-character-1sht-keanu-v187-1675886090936.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s18", 20, 0}},
		},
		{
			"Inside Out 2",
			"Joy, Sadness, Anger, Fear and Disgust have been running a successful operation...",
			"Animation / Family",
			100,
			"https://upload.wikimedia.org/wikipedia/en/f/f7/Inside_Out_2_poster.jpg",
			[]struct {
				ID   string
				Hour int
				Min  int
			}{{"s19", 9, 0}, {"s20", 11, 0}},
		},
	}

	for _, m := range mockMovies {
		var screenings []models.Screening
		for _, st := range m.ScreeningTimes {
			// Create Seats
			var seats []models.Seat
			for r := 0; r < 5; r++ { // 5 Rows
				rowChar := string(rune('A' + r))
				for n := 1; n <= 8; n++ { // 8 Cols
					status := models.SeatAvailable
					seats = append(seats, models.Seat{
						ID:     fmt.Sprintf("%s%d", rowChar, n),
						Row:    rowChar,
						Number: n,
						Status: status,
					})
				}
			}

			screenings = append(screenings, models.Screening{
				ID:        st.ID,
				StartTime: getTime(st.Hour, st.Min),
				Price:     200,
				Seats:     seats,
			})
		}

		newMovie := models.Movie{
			Title:       m.Title,
			Description: m.Description,
			Genre:       m.Genre,
			DurationMin: m.DurationMin,
			PosterURL:   m.PosterURL,
			Screenings:  screenings,
		}

		// Use Upsert to either Insert or Update existing by title
		filter := bson.M{"title": m.Title}
		update := bson.M{
			"$set": bson.M{
				"description":  newMovie.Description,
				"genre":        newMovie.Genre,
				"duration_min": newMovie.DurationMin,
				"poster_url":   newMovie.PosterURL,
				"screenings":   newMovie.Screenings,
			},
		}

		_, err := moviesColl.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
		if err != nil {
			log.Printf("Failed to seed %s: %v", m.Title, err)
		} else {
			log.Printf("Seeded/Updated %s (Price: 200)", m.Title)
		}
	}
	log.Println("All Mock Data Seeded!")
}
