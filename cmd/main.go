package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"bitly/internal/handler"
	"bitly/internal/repository"
	"bitly/internal/service"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MongoURI 		= ""
	DBName       = "url_shortener"
	Port         = ":8081"
	collectionName = "urls"
)

// ensureIndexes sets up necessary indexes for fast lookups and constraint enforcement.
func ensureIndexes(ctx context.Context, client *mongo.Client) {
	collection := client.Database(DBName).Collection(collectionName)
	
	// Index 1: Unique index on ShortCode (crucial for quick lookups and uniqueness check)
	shortCodeIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "shortCode", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Index 2: Index on original URL (crucial for quickly checking if a URL is already shortened)
	urlIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "url", Value: 1}},
	}
	
	// Create the indexes
	names, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{shortCodeIndex, urlIndex})
	if err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}
	log.Printf("Successfully created indexes: %v", names)
}

func main() {
	// 1. Initialize MongoDB Client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the primary to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Successfully connected to MongoDB!")

	// 2. Ensure Database Indexes are present
	ensureIndexes(ctx, client)

	// 3. Initialize Layers
	repo := repository.NewMongoRepository(client, DBName)
	svc := service.NewShortenerService(repo)
	h := handler.NewHandler(svc)

	// 4. Start Server
	router := h.Router()
	
	log.Printf("Starting URL Shortener API on http://localhost%s", Port)
	log.Fatal(http.ListenAndServe(Port, router))

	// Graceful shutdown (optional, but good practice)
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()
}

