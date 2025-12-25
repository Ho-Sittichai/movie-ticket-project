package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	DBURI        string `mapstructure:"DB_URI"`
	DBName       string `mapstructure:"DB_NAME"`
	RedisAddr    string `mapstructure:"REDIS_ADDR"`
	KafkaBrokers string `mapstructure:"KAFKA_BROKERS"`
	Port         string `mapstructure:"PORT"`
}

var AppConfig Config

func LoadConfig() {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_URI", "mongodb://localhost:27017")
	viper.SetDefault("DB_NAME", "movie_ticket_db")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")

	viper.AutomaticEnv()

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode config, %v", err)
	}

	// Manual fallback check
	if AppConfig.DBURI == "" {
		AppConfig.DBURI = os.Getenv("DB_URI")
	}
	if AppConfig.DBName == "" {
		AppConfig.DBName = os.Getenv("DB_NAME")
	}
	if AppConfig.RedisAddr == "" {
		AppConfig.RedisAddr = os.Getenv("REDIS_ADDR")
	}
	if AppConfig.KafkaBrokers == "" {
		AppConfig.KafkaBrokers = os.Getenv("KAFKA_BROKERS")
	}

	log.Println("Config loaded")
}
