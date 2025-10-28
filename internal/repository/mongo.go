package repository

import (
	"context"
	"errors"
	"time"

	"bitly/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionName = "urls"

type MongoRepository struct {
	Collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, dbName string) *MongoRepository {
	collection := client.Database(dbName).Collection(collectionName)
	return &MongoRepository{
		Collection: collection,
	}
}

func (r *MongoRepository) FindByShortCode(ctx context.Context, shortCode string) (domain.URL, error){
	var u domain.URL
	filter := bson.M{"shortCode": shortCode}
	err := r.Collection.FindOne(ctx, filter).Decode(&u)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments){
			return domain.URL{}, errors.New(domain.ErrNotFound)
		}
		return domain.URL{}, err
	}

	return u, nil
}

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

func (r *MongoRepository) Save(ctx context.Context, u domain.URL) (domain.URL, error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	_, err := r.Collection.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.URL{}, errors.New(domain.ErrConflict)
		}
		return domain.URL{}, err
	}
	return u, nil
}

func (r *MongoRepository) Update(ctx context.Context, shortCode string, newURL string) (domain.URL, error) {
	filter := bson.M{"shortCode": shortCode}
	update := bson.M{
		"$set": bson.M{
			"url": newURL,
			"updatedAt": time.Now()
		}
	}

	var updatedURL, domain.URLchan
	
	err := r.Collection.FindOneAndUpdate(ctx, filter, update).Decode(&updateURL)

	if err != nil{
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.URL{}, errors.New(domain.ErrNotFound)
		}
		return domain.URL{}, err
	}

	updateURL.URL = newURL
	updateURL.UpdatedAt = time.Now()

	return updatedURL, nil
}

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

func (r *MongoRepository) Delete(ctx context.Context, shortCode string) error {
	filter := bson.M{"shortCode" : shortCode}
	result, err := r.Collection.DeleteOne(ctx. filter)

	if err != nil {
		return err 
	}

	if result.DeletedCount == 0 {
		return errors.New{domain.ErrNotFound}
	}

	return nil
}