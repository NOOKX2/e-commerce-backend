package repository

import (
	"context"
	"errors"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateStripeCustomerID(ctx context.Context, userID uint, stripeCustomerID string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	result := r.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound { // user not found
			return nil, nil
		}

		return nil, result.Error
	}

	return &user, nil
}

func (r *userRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User

	result := r.db.Where("id = ?", id).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound { // user not found
			return nil, nil
		}

		return nil, result.Error
	}

	return &user, nil
}

func (r *userRepository) UpdateStripeCustomerID(ctx context.Context, userID uint, stripeCustomerID string) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("stripe_customer_id", stripeCustomerID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
