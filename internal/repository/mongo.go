package repository

import (
	"context"
	"errors"
	"time"

	"github.com/hahahamid/broker-backend/config"
	"github.com/hahahamid/broker-backend/internal/models"
	"github.com/hahahamid/broker-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/sony/gobreaker"
	"golang.org/x/crypto/bcrypt"
)

type MongoRepo struct {
	client *mongo.Client
	db     *mongo.Database
	userCB *gobreaker.CircuitBreaker
}

func NewMongoRepo(cfg *config.Config) (*MongoRepo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &MongoRepo{
		client: client,
		db:     client.Database(cfg.DBName),
		userCB: utils.NewCB("mongo-users"),
	}, nil
}

func (r *MongoRepo) CreateUser(ctx context.Context, email, password string) error {
	pwHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{Email: email, PasswordHash: string(pwHash)}

	_, err := r.userCB.Execute(func() (interface{}, error) {
		return r.db.Collection("users").InsertOne(ctx, user)
	})
	if mongo.IsDuplicateKeyError(err) {
		return errors.New("email already exists")
	}
	return err
}

func (r *MongoRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	res, err := r.userCB.Execute(func() (interface{}, error) {
		// return both value and error
		return r.db.Collection("users").FindOne(ctx, bson.M{"email": email}), nil
	})
	if err != nil {
		return nil, err
	}

	singleResult := res.(*mongo.SingleResult)
	if err := singleResult.Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoRepo) SaveRefreshToken(ctx context.Context, userID, token string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = r.db.Collection("users").UpdateByID(ctx, oid, bson.M{"$set": bson.M{"refresh_token": token}})
	return err
}
