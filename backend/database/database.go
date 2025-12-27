package database

import (
	"context"
	"log"
	"movie-ticket-backend/config"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Mongo *mongo.Database
	RDB   *redis.Client
)

func ConnectDB() {
	// 1. Connect MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.AppConfig.DBURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
	} else {
		err = client.Ping(ctx, nil)
		if err != nil {
			log.Printf("Failed to ping MongoDB: %v", err)
		} else {
			log.Println("Connected to MongoDB successfully")
			Mongo = client.Database(config.AppConfig.DBName)
		}
	}

	// 2. Connect Redis
	RDB = redis.NewClient(&redis.Options{
		Addr: config.AppConfig.RedisAddr,
	})

	_, err = RDB.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		log.Println("Connected to Redis successfully")
		// Enable keyspace notifications for expired events
		RDB.ConfigSet(context.Background(), "notify-keyspace-events", "Ex")
	}
}
