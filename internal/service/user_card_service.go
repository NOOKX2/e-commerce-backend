package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"gorm.io/gorm"
)

type UserCardServiceInterface interface {
	SaveNewCard(ctx context.Context, userID uint, req request.SaveCardRequest) (uint, error)
	GetCardByUserID(ctx context.Context, userID uint) ([]models.UserCard, error)
	GetCardByID(ctx context.Context, userID uint, cardID uint) (*models.UserCard, error)
	GetCardByStripePaymentMethodID(ctx context.Context, userID uint, stripePaymentMethodID string) (*models.UserCard, error)
	DeleteCard(ctx context.Context, cardID uint, userID uint) error
}

type UserCardService struct {
	repo repository.UserCardRepositoryInterface
}

func NewUserCardService(repo repository.UserCardRepositoryInterface) UserCardServiceInterface {
	return &UserCardService{repo: repo}
}

func (s *UserCardService) SaveNewCard(ctx context.Context, userID uint, req request.SaveCardRequest) (uint, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	stripePM, err := paymentmethod.Get(req.PaymentMethodID, nil)
	if err != nil {
		return 0, err
	}

	uniqueKey := stripePM.Card.Fingerprint
	existingCard, err := s.repo.GetCardByUniqueKey(ctx, userID, uniqueKey)
	if err == nil {
		return existingCard.ID, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
		stripePM, err := paymentmethod.Get(req.PaymentMethodID, nil)
		if err != nil {
			return 0, fmt.Errorf("could not retrieve payment method from stripe: %w", err)
		}
		card := &models.UserCard{
			UserID:                userID,
			StripePaymentMethodID: req.PaymentMethodID,
			CardBrand:             string(stripePM.Card.Brand),
			LastFour:              stripePM.Card.Last4,
			ExpiryMonth:           int(stripePM.Card.ExpMonth),
			ExpiryYear:            int(stripePM.Card.ExpYear),
			IsDefault:             false,
		}

		if err := s.repo.Create(ctx, card); err != nil {
			return 0, err
		}
		return card.ID, nil
	}

	return 0, err
}

func (s *UserCardService) GetCardByUserID(ctx context.Context, userID uint) ([]models.UserCard, error) {
	cards, err := s.repo.GetCardByUserID(ctx, userID)

	if err != nil {
		return nil, err
	}

	return cards, nil
}

func (s *UserCardService) GetCardByID(ctx context.Context, userID uint, cardID uint) (*models.UserCard, error) {
	card, err := s.repo.GetCardByID(ctx, userID, cardID)
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (s *UserCardService) GetCardByStripePaymentMethodID(ctx context.Context, userID uint, stripePaymentMethodID string) (*models.UserCard, error) {
	card, err := s.repo.GetCardByStripePaymentMethodID(ctx, userID, stripePaymentMethodID)
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (s *UserCardService) DeleteCard(ctx context.Context, cardID uint, userID uint) error {
	card, err := s.repo.GetCardByID(ctx, userID, cardID)
	if err != nil {
		return fmt.Errorf("card not found")
	}

	_, err = paymentmethod.Detach(card.StripePaymentMethodID, nil)
	if err != nil {
		log.Printf("Stripe Detach Warning: %v", err)
	}

	if err = s.repo.DeleteCard(ctx, cardID, userID); err != nil {
		return fmt.Errorf("could not delete card from database: %v", err)
	}

	return nil
}
