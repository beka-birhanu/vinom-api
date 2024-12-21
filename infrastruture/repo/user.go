package repo

import (
	"context"
	"errors"
	"time"

	dmn "github.com/beka-birhanu/vinom-api/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepo handles the persistence of user models.
type UserRepo struct {
	collection *mongo.Collection
}

// NewUserRepo creates a new UserRepo with the given MongoDB client, database name, and collection name.
func NewUserRepo(client *mongo.Client, dbName, collectionName string) *UserRepo {
	collection := client.Database(dbName).Collection(collectionName)
	return &UserRepo{
		collection: collection,
	}
}

// Save inserts or updates a user in the repository.
// If the user already exists, it updates the existing record.
// If the user does not exist, it adds a new record.
func (u *UserRepo) Save(user *dmn.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"username":     user.Username,
			"passwordHash": user.PasswordHash,
			"rating":       user.Rating,
			"updatedAt":    time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := u.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("username conflict")
		}
		return errors.New("unexpected error: " + err.Error())
	}

	return nil
}

// ByID retrieves a user by their ID.
// Returns an error if the user is not found or if an unexpected error occurs.
func (u *UserRepo) ByID(id uuid.UUID) (*dmn.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	var user dmn.User
	if err := u.collection.FindOne(ctx, filter).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("unexpected error: " + err.Error())
	}
	return &user, nil
}

// ByUsername retrieves a user by their username.
// Returns an error if the user is not found or if an unexpected error occurs.
func (u *UserRepo) ByUsername(username string) (*dmn.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	filter := bson.M{"username": username}
	var user dmn.User
	if err := u.collection.FindOne(ctx, filter).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("unexpected error: " + err.Error())
	}
	return &user, nil
}
