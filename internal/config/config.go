package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Lets construct a struct of type config
type Config struct {
	DB_HOST string
	DB_NAME string
	DB_USER string
	DB_PASS string
	DB_PORT string
	JWTSecret string
    JWTExpiry string
}

func Load() *Config {
	log.Println("Loading Environment variables")
	godotenv.Load()
	
	return &Config{
		DB_HOST: os.Getenv("DB_HOST"),
		DB_NAME: os.Getenv("DB_NAME"),
		DB_USER: os.Getenv("DB_USER"),
		DB_PASS: os.Getenv("DB_PASS"),
		DB_PORT: os.Getenv("DB_PORT"),
		JWTSecret: os.Getenv("JWT_SECRET"),
    	JWTExpiry: os.Getenv("JWT_EXPIRY"),
	}
}