package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	ServiceID      string
	ServiceSecret  string
	AuthServiceURL string
	JWTSecret      string
)

func LoadEnv() {
	_ = godotenv.Load() // silently load .env if present

	ServiceID = os.Getenv("SERVICE_ID")
	ServiceSecret = os.Getenv("SERVICE_SECRET")
	AuthServiceURL = os.Getenv("AUTH_SERVICE_URL")
	JWTSecret = os.Getenv("JWT_SECRET")

	if ServiceID == "" || ServiceSecret == "" || AuthServiceURL == "" {
		log.Fatal("Missing required environment variables")
	}

	// JWT_SECRET is optional for services that don't need local JWT validation
	if JWTSecret == "" {
		log.Print("WARNING: JWT_SECRET not set. Local JWT validation will not be available.")
	}
}
