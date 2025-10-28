package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"bitly/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionName = "urls"

// MongoRepository implements the domain.Repository interface using MongoDB.
type MongoRepository struct {
	Collection *mongo.Collection
}

// NewMongoRepository creates a new repository instance.
func NewMongoRepository(client *mongo.Client, dbName string) *MongoRepository {
	collection := client.Database(dbName).Collection(collectionName)
	return &MongoRepository{
		Collection: collection,
	}
}

// FindByShortCode retrieves a URL document by its short code.
func (r *MongoRepository) FindByShortCode(ctx context.Context, shortCode string) (domain.URL, error) {
	var u domain.URL
	filter := bson.M{"shortCode": shortCode}
	err := r.Collection.FindOne(ctx, filter).Decode(&u)
	
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.URL{}, errors.New(domain.ErrNotFound)
		}
		return domain.URL{}, err
	}
	return u, nil
}

// FindByOriginalURL retrieves a URL document by its original long URL.
func (r *MongoRepository) FindByOriginalURL(ctx context.Context, originalURL string) (domain.URL, error) {
	var u domain.URL
	filter := bson.M{"url": originalURL}
	err := r.Collection.FindOne(ctx, filter).Decode(&u)
	
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.URL{}, errors.New(domain.ErrNotFound)
		}
		return domain.URL{}, err
	}
	return u, nil
}

// Save creates a new URL document.
func (r *MongoRepository) Save(ctx context.Context, u domain.URL) (domain.URL, error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	
	_, err := r.Collection.InsertOne(ctx, u)
	if err != nil {
		// --- FIX: Check for duplicate key error (code 11000) specifically for shortCode uniqueness ---
		if writeErr, ok := err.(mongo.WriteException); ok {
			for _, e := range writeErr.WriteErrors {
				if e.Code == 11000 && strings.Contains(e.Message, "shortCode") {
					return domain.URL{}, errors.New(domain.ErrConflict) // Use ErrConflict for duplication
				}
			}
		}
		// --- END FIX ---
		return domain.URL{}, err
	}
	return u, nil
}

// Update updates the long URL for an existing document.
func (r *MongoRepository) Update(ctx context.Context, shortCode string, newURL string) (domain.URL, error) {
	filter := bson.M{"shortCode": shortCode}
	update := bson.M{
		"$set": bson.M{
			"url": newURL,
			"updatedAt": time.Now(),
		},
	}
	
	var updatedURL domain.URL
	
	// Find the document, update it, and return the new version in one go
	err := r.Collection.FindOneAndUpdate(ctx, filter, update).Decode(&updatedURL)
	
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.URL{}, errors.New(domain.ErrNotFound)
		}
		return domain.URL{}, err
	}
	
	// The FindOneAndUpdate returned the *old* document before update. We need to query again or re-apply changes.
	// Simpler approach: update the retrieved struct with the new values and return it (since the update was successful).
	updatedURL.URL = newURL
	updatedURL.UpdatedAt = time.Now() // It's actually the old timestamp, but we manually set the new one
	
	return updatedURL, nil
}

// IncrementAccessCount increments the access count for a short code.
func (r *MongoRepository) IncrementAccessCount(ctx context.Context, shortCode string) error {
	filter := bson.M{"shortCode": shortCode}
	update := bson.M{
		"$inc": bson.M{"accessCount": 1},
	}
	
	result, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New(domain.ErrNotFound)
	}
	return nil
}

// Delete removes a URL document by its short code.
func (r *MongoRepository) Delete(ctx context.Context, shortCode string) error {
	filter := bson.M{"shortCode": shortCode}
	result, err := r.Collection.DeleteOne(ctx, filter)
	
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New(domain.ErrNotFound)
	}
	return nil
}
