package repository

import (
	"context"

	"github.com/hahahamid/broker-backend/internal/models"
)

type UserRepo interface {
	CreateUser(ctx context.Context, email, password string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	SaveRefreshToken(ctx context.Context, userID, token string) error
}
