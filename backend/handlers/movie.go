package handlers

import (
	"context"
	"movie-ticket-backend/database"
	"movie-ticket-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Movie Handler ---
func GetMovies(c *gin.Context) {
	collection := database.Mongo.Collection("movies")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var movies []models.Movie
	if err = cursor.All(context.TODO(), &movies); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Remove seats from all screenings in the list
	for i := range movies {
		for j := range movies[i].Screenings {
			movies[i].Screenings[j].Seats = nil
		}
	}

	c.JSON(http.StatusOK, movies)
}

func CreateMovie(c *gin.Context) {
	var movie models.Movie
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	movie.ID = primitive.NewObjectID()
	collection := database.Mongo.Collection("movies")
	_, err := collection.InsertOne(context.TODO(), movie)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, movie)
}
