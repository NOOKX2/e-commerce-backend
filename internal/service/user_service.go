package service

import (
	"errors"

	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/domain"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
)

type UserServiceInterface interface {
	Register(email, password,name string) (*domain.User, error)
	Login(email, password string) (string, error)
	GetUserByID(id uint) (*domain.User, error)
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

func (s *UserService) Register(email, password,name string) (*domain.User, error) {
	hashedPassword, err := utils.HashPassword(password)

	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	newUser := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:        name,
	}

	err = s.userRepo.Create(newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *UserService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("email not found")
	}

	validPassword := utils.CheckHashedPassword(password, user.PasswordHash)
	if !validPassword {
		return "", errors.New("invalid password")
	}

	token, err := utils.GenerateToken(user, s.config.JWTSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) GetUserByID(id uint) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(id)

	if err != nil {
		return nil, err
	}

	return user, nil
}
