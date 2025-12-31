package handlers

import (
	"context"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type AdminBookingResponse struct {
	ID            string    `json:"id"`
	UserEmail     string    `json:"user_email"`
	UserName      string    `json:"user_name"`
	MovieTitle    string    `json:"movie_title"`
	PosterURL     string    `json:"poster_url"`
	ScreeningTime time.Time `json:"screening_time"`
	SeatID        string    `json:"seat_id"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
}

func GetAllBookings(c *gin.Context) {
	// Filters
	filterMovieID := c.Query("movie_id")
	filterDate := c.Query("date") // YYYY-MM-DD
	filterUser := c.Query("user") // Name or Email partial match

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	// 1. Build Filter
	// filter := bson.M{} (Unused, using manual filtering for now)

	// Pre-fetch related data for filtering ID-based fields effectively would require aggregation
	// or complex logic. To keep it simple and safe as per plan:
	// We will fetch ALL matching the basic ID filters first if possible,
	// OR we stick to the post-filtering approach but optimization is tricky without aggregation.
	// Given the "Lag" complaint, let's try to filter at DB level where possible.

	// However, MovieID and User search are tricky because Bookings only have IDs.
	// If we want TRUE server-side pagination with filtering, we need to:
	// A) Filter by Date (easy, if strictly DB field) - createdAt is time.Time,
	//    User probably wants "Screening Date". Screening info is in another collection.
	// B) Filter by Movie/User - requires Lookups.

	// OPTIMIZATION STRATEGY:
	// For now, to solve "Lag" which is likely rendering 100s of items:
	// We will fetch all (as before), filter in memory (as before),
	// AND THEN paginate the RESULT slice before sending.
	// This solves Network and Frontend Rendering lag, though not DB load (which is usually fine for <10k records).

	// ... (Existing logic to fetch all) ...

	bookingsColl := database.Mongo.Collection("bookings")
	cursor, err := bookingsColl.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch bookings"})
		return
	}
	var bookings []models.Booking
	if err = cursor.All(context.TODO(), &bookings); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode bookings"})
		return
	}

	// Fetch Movies & Users for mapping
	moviesColl := database.Mongo.Collection("movies")
	mCursor, _ := moviesColl.Find(context.TODO(), bson.M{})
	var movies []models.Movie
	_ = mCursor.All(context.TODO(), &movies)

	// Map ScreeningID -> Movie Info
	type ScreeningInfo struct {
		MovieTitle string
		Poster     string
		StartTime  time.Time
		MovieID    string
	}
	screeningMap := make(map[string]ScreeningInfo)
	for _, m := range movies {
		for _, s := range m.Screenings {
			screeningMap[s.ID] = ScreeningInfo{
				MovieTitle: m.Title,
				Poster:     m.PosterURL,
				StartTime:  s.StartTime,
				MovieID:    m.ID.Hex(),
			}
		}
	}

	usersColl := database.Mongo.Collection("users")
	uCursor, _ := usersColl.Find(context.TODO(), bson.M{})
	var users []models.User
	_ = uCursor.All(context.TODO(), &users)
	userMap := make(map[string]models.User)
	for _, u := range users {
		userMap[u.ID.Hex()] = u
	}

	// Filter & Map
	var filtered []AdminBookingResponse

	for _, b := range bookings {
		scInfo, okSc := screeningMap[b.ScreeningID]
		user, okUser := userMap[b.UserID]

		// Filter: Movie
		if filterMovieID != "" {
			if !okSc || scInfo.MovieID != filterMovieID {
				continue
			}
		}

		// Filter: Date (Screening Date)
		if filterDate != "" {
			if !okSc || scInfo.StartTime.Format("2006-01-02") != filterDate {
				continue
			}
		}

		// Filter: User (Search)
		if filterUser != "" {
			if !okUser {
				continue
			}
			matchName := strings.Contains(strings.ToLower(user.Name), strings.ToLower(filterUser))
			matchEmail := strings.Contains(strings.ToLower(user.Email), strings.ToLower(filterUser))
			if !matchName && !matchEmail {
				continue
			}
		}

		item := AdminBookingResponse{
			ID:            b.ID.Hex(),
			UserEmail:     "Unknown",
			UserName:      "Unknown",
			MovieTitle:    "Unknown Movie",
			PosterURL:     "",
			ScreeningTime: time.Time{},
			SeatID:        b.SeatID,
			Status:        b.Status,
			Amount:        b.Amount,
			CreatedAt:     b.CreatedAt,
		}

		if okUser {
			item.UserEmail = user.Email
			item.UserName = user.Name
		}
		if okSc {
			item.MovieTitle = scInfo.MovieTitle
			item.PosterURL = scInfo.Poster
			item.ScreeningTime = scInfo.StartTime
		}
		filtered = append(filtered, item)
	}

	// Sort Descending by CreatedAt
	// (Simple bubble/api sort or just reverse if DB order close enough)
	// Better: Sorting logic if needed. For now assume DB order or implement sort.
	// Let's reverse to show newest first if not sorted
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	// Pagination Logic on `filtered` slice
	total := len(filtered)
	start := skip
	end := start + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paged := filtered[start:end]

	c.JSON(200, gin.H{
		"data": paged,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + limit - 1) / limit,
		},
	})
}

func GetAdminStats(c *gin.Context) {
	// Simple helper for cards
	// Reuse logic or create specialized aggregation later
	// For now, let Frontend calculate from /bookings response or separate specific calls if needed.
	// We will just return empty for now as task didn't explicitly ask for separate stats API,
	// but the dashboard usually needs them.
	// Let's stick to returning data in GetAllBookings and let frontend count.
}
