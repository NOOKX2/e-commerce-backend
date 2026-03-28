package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateStripeCustomerID(ctx context.Context, userID uint, stripeCustomerID string) error
	UpdateProfile(userID uint, name, email string) error
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

		fmt.Println("error here", result.Error);

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

func (r *userRepository) UpdateProfile(userID uint, name, email string) error {
	var count int64
	if err := r.db.Model(&models.User{}).Where("email = ? AND id != ?", email, userID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("email already in use")
	}

	result := r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"name":  name,
		"email": email,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
