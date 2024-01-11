package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

func NewConfig() *Config {
	var c Config
	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&c.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.StringVar(&c.FileStoragePath, "f", "", "Path for storage file")
	flag.Parse()

	if envServerAddress, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		c.ServerAddress = envServerAddress
	}
	if envBaseURL, exists := os.LookupEnv("BASE_URL"); exists {
		c.BaseURL = envBaseURL
	}
	if envFileStoragePage, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		c.FileStoragePath = envFileStoragePage
	}

	return &c
}
