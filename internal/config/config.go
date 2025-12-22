package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	// LinkedIn credentials
	LinkedInEmail    string
	LinkedInPassword string

	// Browser settings
	Headless   bool
	SlowMotion int // milliseconds

	// Database
	DatabasePath string

	// Session
	SessionFile string

	// Rate limiting
	MaxConnectionsPerDay int
	DelayBetweenActions  int // seconds

	// Search settings
	SearchQuery   string
	SearchFilters string
	MaxResults    int
}

// LoadConfig loads configuration from environment variables
func LoadConfig(envFile string) (*Config, error) {
	// Load .env file if it exists
	if envFile == "" {
		envFile = ".env"
	}

	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
		fmt.Printf("Loaded configuration from: %s\n", envFile)
	}

	config := &Config{
		LinkedInEmail:        getEnv("LINKEDIN_EMAIL", ""),
		LinkedInPassword:     getEnv("LINKEDIN_PASSWORD", ""),
		Headless:             getEnvBool("HEADLESS", false),
		SlowMotion:           getEnvInt("SLOW_MOTION", 0),
		DatabasePath:         getEnv("DATABASE_PATH", "linkedin_bot.db"),
		SessionFile:          getEnv("SESSION_FILE", "session.json"),
		MaxConnectionsPerDay: getEnvInt("MAX_CONNECTIONS_PER_DAY", 50),
		DelayBetweenActions:  getEnvInt("DELAY_BETWEEN_ACTIONS", 5),
		SearchQuery:          getEnv("SEARCH_QUERY", ""),
		SearchFilters:        getEnv("SEARCH_FILTERS", ""),
		MaxResults:           getEnvInt("MAX_RESULTS", 100),
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.LinkedInEmail == "" {
		return fmt.Errorf("LINKEDIN_EMAIL is required")
	}

	if c.LinkedInPassword == "" {
		return fmt.Errorf("LINKEDIN_PASSWORD is required")
	}

	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
