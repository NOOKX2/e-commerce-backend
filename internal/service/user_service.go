package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
)

type UserServiceInterface interface {
	Register(email, password, name string) (string, *models.User, error)
	Login(email, password string) (string, *models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateStripeCustomerID(ctx context.Context, userID uint, stripeCustomerID string) error
}

type UserService struct {
	userRepo repository.UserRepositoryInterface
	config   *configs.Config
}

func NewUserService(repo repository.UserRepositoryInterface, cfg *configs.Config) UserServiceInterface {
	return &UserService{
		userRepo: repo,
		config:   cfg,
	}
}

func (s *UserService) Register(email, password, name string) (string, *models.User, error) {
	existedUser, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", nil, err
	}

	if existedUser != nil {
		fmt.Println("request come here in service second")
		return "", nil, ErrUserExisted
	}

	hashedPassword, err := utils.HashPassword(password)

	if err != nil {
		return "", nil, errors.New("failed to hash password")
	}

	newUser := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
	}

	err = s.userRepo.Create(newUser)
	if err != nil {
		return "", nil, err
	}

	token, err := utils.GenerateToken(newUser, s.config.JWTSecret)
	if err != nil {
		return "", nil, err
	}

	return token, newUser, nil
}

func (s *UserService) Login(email, password string) (string, *models.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, ErrUserNotFound
	}

	validPassword := utils.CheckHashedPassword(password, user.PasswordHash)
	if !validPassword {
		return "", nil, ErrPasswordIncorrect
	}

	token, err := utils.GenerateToken(user, s.config.JWTSecret)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateStripeCustomerID(ctx context.Context, userID uint, stripeCustomerID string) error {
	if err := s.userRepo.UpdateStripeCustomerID(ctx, userID, stripeCustomerID); err != nil {
		return fmt.Errorf("could not update stripe customer id: %w", err)
	}

	return nil
}
