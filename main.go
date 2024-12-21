package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	redisClient *redis.Client
	mongoClient *mongo.Client
)

func initRedis(ctx context.Context) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		DB:   0,
	})

	// Test Redis connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis!")
}

func initMongo(ctx context.Context) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", dbUser, dbPass, dbHost, dbPort)
	clientOptions := options.Client().ApplyURI(uri)

	var err error
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Test MongoDB connection
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}
	log.Println("Connected to MongoDB!")
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize connections
	initRedis(ctx)
	initMongo(ctx)

	// Example Redis command
	val, err := redisClient.Get(ctx, "key").Result()
	if err == redis.Nil {
		log.Println("key does not exist")
	} else if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	} else {
		log.Printf("key: %s", val)
	}

	// Example MongoDB command
	collection := mongoClient.Database("test").Collection("example")
	_, err = collection.InsertOne(ctx, map[string]string{"message": "Hello, MongoDB!"})
	if err != nil {
		log.Fatalf("Failed to insert into MongoDB: %v", err)
	} else {
		log.Println("Document inserted into MongoDB!")
	}

	// Close clients on exit
	defer func() {
		_ = redisClient.Close()
		_ = mongoClient.Disconnect(ctx)
	}()
}
