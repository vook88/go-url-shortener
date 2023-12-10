package config

import (
	"flag"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() *Config {
	serverAddress := flag.String("a", "localhost:8080", "HTTP server address")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL for shortened URLs")

	flag.Parse()

	return &Config{
		ServerAddress: *serverAddress,
		BaseURL:       *baseURL,
	}
}
