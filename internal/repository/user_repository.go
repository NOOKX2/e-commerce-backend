package repository

import (
	"github.com/NOOKX2/e-commerce-backend/internal/domain"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(user *domain.User) error
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByID(id uint) (*domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r *userRepository) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User

	result := r.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound { // user not found
			return nil, nil
		}

		return nil, result.Error
	}

	return &user, nil
}

func (r *userRepository) GetUserByID(id uint) (*domain.User, error) {
	var user domain.User

	result := r.db.Where("id = ?", id).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound { // user not found
			return nil, nil
		}

		return nil, result.Error
	}

	return &user, nil
}
