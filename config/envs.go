package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration values.
type Config struct {
	HostIP             string // Host IP for the server
	RESTPort           int    // Port for the REST API
	DBHost             string // Hostname or IP address for the database
	DBPort             int    // Port number for the database
	DBUser             string // Username for the database
	DBPassword         string // Password for the database
	DBName             string // Name of the database
	MatchmakingHost    string // Hostname or IP address for the Matchmaiking server
	MatchmakingPort    int    // Port number for the Matchmaiking server
	SessionManagerHost string // Hostname or IP address for the session manager server
	SessionManagerPort int    // Port number for the session manager server
	GinMode            string // Mode for the Gin framework (e.g., release, debug, test)
	JWTSecret          string // Secret key for JWT signing
	JWTIssuer          string // Issuer claim for JWTs
}

// Envs holds the application's configuration loaded from environment variables.
var Envs = initConfig()

// initConfig initializes and returns the application configuration.
// It loads environment variables from a .env file.
func initConfig() Config {
	// Load .env file if available
	if err := godotenv.Load(); err != nil {
		log.Printf("[APP] [INFO] .env file not found or could not be loaded: %v", err)
	}

	// Populate the Config struct with required environment variables
	return Config{
		DBHost:             mustGetEnv("DB_HOST"),
		DBPort:             mustGetEnvAsInt("DB_PORT"),
		DBUser:             mustGetEnv("DB_USER"),
		DBPassword:         mustGetEnv("DB_PASS"),
		DBName:             mustGetEnv("DB_NAME"),
		MatchmakingHost:    mustGetEnv("MATCHMAKING_HOST"),
		MatchmakingPort:    mustGetEnvAsInt("MATCHMAKING_PORT"),
		SessionManagerHost: mustGetEnv("SESSION_HOST"),
		SessionManagerPort: mustGetEnvAsInt("SESSION_PORT"),
		GinMode:            getEnvWithDefault("GIN_MODE", "release"),
		JWTSecret:          mustGetEnv("JWT_SECRET"),
		JWTIssuer:          mustGetEnv("JWT_ISSUER"),
		HostIP:             mustGetEnv("HOST_IP"),
		RESTPort:           mustGetEnvAsInt("REST_PORT"),
	}
}

// mustGetEnv retrieves the value of an environment variable or logs a fatal error if not set.
func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("[APP] [FATAL] Environment variable %s is not set", key)
	}
	return value
}

// mustGetEnvAsInt retrieves the value of an environment variable as an integer or logs a fatal error if not set or cannot be parsed.
func mustGetEnvAsInt(key string) int {
	valueStr := mustGetEnv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("[APP] [FATAL] Environment variable %s must be an integer: %v", key, err)
	}
	return value
}

// getEnvWithDefault retrieves the value of an environment variable or returns a default value if not set.
func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
