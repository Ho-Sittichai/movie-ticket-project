package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	DBURI              string `mapstructure:"DB_URI"`
	DBName             string `mapstructure:"DB_NAME"`
	RedisAddr          string `mapstructure:"REDIS_ADDR"`
	KafkaBrokers       string `mapstructure:"KAFKA_BROKERS"`
	Port               string `mapstructure:"PORT"`
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`
}

var AppConfig Config

func LoadConfig() {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_URI", "mongodb://localhost:27017")
	viper.SetDefault("DB_NAME", "movie_ticket_db")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback")

	// Load from .env file if it exists
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	viper.AutomaticEnv()

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode config, %v", err)
	}

	if AppConfig.GoogleClientID == "" {
		AppConfig.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	}
	if AppConfig.GoogleClientSecret == "" {
		AppConfig.GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	}
	if AppConfig.GoogleRedirectURL == "" {
		AppConfig.GoogleRedirectURL = os.Getenv("GOOGLE_REDIRECT_URL")
	}

	log.Println("Config loaded")
}
