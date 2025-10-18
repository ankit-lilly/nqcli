package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	URL   string
	Token string
}

func LoadConfig() *Config {
	godotenv.Load()

	cfg := &Config{
		URL:   os.Getenv("NEPTUNE_URL"),
		Token: os.Getenv("NEPTUNE_TOKEN"),
	}

	if cfg.URL == "" {
		cfg.URL = "https://ljoq4kni2bel7c2hab7dpdk45y.appsync-api.us-east-2.amazonaws.com/graphql"
	}
	if cfg.Token == "" {
		cfg.Token = "jwt token"
	}

	if cfg.Token == "jwt token" {
		log.Println("Warning: Using placeholder 'jwt token'. Set NEPTUNE_TOKEN environment variable.")
	}

	return cfg
}
