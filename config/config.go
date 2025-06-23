package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI             string
	DBName               string
	JWTSecret            string
	RefreshSecret        string
	AccessTokenExpireMin int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	exp, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRE_MINUTES"))
	if err != nil {
		exp = 10
	}

	return &Config{
		MongoURI:             os.Getenv("MONGO_URI"),
		DBName:               os.Getenv("DB_NAME"),
		JWTSecret:            os.Getenv("JWT_SECRET"),
		RefreshSecret:        os.Getenv("REFRESH_SECRET"),
		AccessTokenExpireMin: exp,
	}
}
