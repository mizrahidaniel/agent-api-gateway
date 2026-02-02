package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the gateway configuration
type Config struct {
	Upstream struct {
		URL string `yaml:"url"`
	} `yaml:"upstream"`
	Auth struct {
		Provider string `yaml:"provider"`
		Keys     []struct {
			Key       string `yaml:"key"`
			Tier      string `yaml:"tier"`
			RateLimit string `yaml:"rate_limit"`
		} `yaml:"keys"`
	} `yaml:"auth"`
	Analytics struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"analytics"`
	Monitoring struct {
		HealthCheck string `yaml:"health_check"`
		Interval    string `yaml:"interval"`
	} `yaml:"monitoring"`
}

func main() {
	configFile := flag.String("config", "gateway.yaml", "Configuration file path")
	port := flag.Int("port", 8080, "Gateway port")
	flag.Parse()

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Agent API Gateway v0.1.0\n")
	fmt.Printf("Upstream: %s\n", config.Upstream.URL)
	fmt.Printf("Port: %d\n", *port)
	fmt.Printf("API Keys: %d configured\n", len(config.Auth.Keys))

	// TODO: Start HTTP server
	// TODO: Implement proxy handler
	// TODO: Add authentication middleware
	// TODO: Add rate limiting middleware
	// TODO: Add logging middleware

	fmt.Println("\nMVP in progress - coming soon!")
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &config, nil
}
