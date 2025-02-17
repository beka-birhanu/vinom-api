package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration values.
type Config struct {
	DBHost                 string // Hostname or IP address for the database
	DBPort                 int    // Port number for the database
	DBUser                 string // Username for the database
	DBPassword             string // Password for the database
	DBName                 string // Name of the database
	MatchmakingHost        string // Hostname or IP address for the Matchmaiking server
	MatchmakingPort        int    // Port number for the Matchmaiking server
	SessionManagerHost     string // Hostname or IP address for the session manager server
	SessionManagerPort     int    // Port number for the session manager server
	GinMode                string // Mode for the Gin framework (e.g., release, debug, test)
	UDPBufferSize          int    // Size of the buffer for incoming UDP packets (in bytes)
	UDPHeartbeatExpiration int    // Expiration time for UDP heartbeat (in milliseconds)
	JWTSecret              string // Secret key for JWT signing
	JWTIssuer              string // Issuer claim for JWTs
	MaxPlayer              int    // Maximum number of players allowed in a game
	RankTolerance          int    // Tolerance for player rank difference during matchmaking
	LatencyTolerance       int    // Tolerance for latency (in milliseconds) during matchmaking
	HostIP                 string // Host IP for the server
	RESTPort               int    // Port for the REST API
	UDPPort                int    // Port for the UDP server
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
		DBHost:                 mustGetEnv("DB_HOST"),
		DBPort:                 mustGetEnvAsInt("DB_PORT"),
		DBUser:                 mustGetEnv("DB_USER"),
		DBPassword:             mustGetEnv("DB_PASS"),
		DBName:                 mustGetEnv("DB_NAME"),
		MatchmakingHost:        mustGetEnv("MATCHMAKING_HOST"),
		MatchmakingPort:        mustGetEnvAsInt("MATCHMAKING_PORT"),
		SessionManagerHost:     mustGetEnv("SESSION_HOST"),
		SessionManagerPort:     mustGetEnvAsInt("SESSION_PORT"),
		GinMode:                getEnvWithDefault("GIN_MODE", "release"),
		UDPBufferSize:          mustGetEnvAsInt("UDP_BUFFER_SIZE"),
		UDPHeartbeatExpiration: mustGetEnvAsInt("UDP_HEARTBEAT_EXPIRATION"),
		JWTSecret:              mustGetEnv("JWT_SECRET"),
		JWTIssuer:              mustGetEnv("JWT_ISSUER"),
		MaxPlayer:              mustGetEnvAsInt("MAX_PLAYER"),
		RankTolerance:          mustGetEnvAsInt("RANK_TOLERANCE"),
		LatencyTolerance:       mustGetEnvAsInt("LATENCY_TOLERANCE"),
		HostIP:                 mustGetEnv("HOST_IP"),
		RESTPort:               mustGetEnvAsInt("REST_PORT"),
		UDPPort:                mustGetEnvAsInt("UDP_PORT"),
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
