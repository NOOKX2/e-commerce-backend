package repository

import (
	"context"
	"fmt"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type userCardRepository struct {
	db *gorm.DB
}

type UserCardRepositoryInterface interface {
	Create(ctx context.Context, card *models.UserCard) error
	GetCardByUserID(ctx context.Context, userID uint) ([]models.UserCard, error)
	CheckCardExists(ctx context.Context, userID uint, paymentMethodID string) (bool, error)
	GetCardByID(ctx context.Context, userID uint, cardID uint) (*models.UserCard, error)
	GetCardByStripePaymentMethodID(ctx context.Context, userID uint, stripePaymentMethodID string) (*models.UserCard, error)
	GetCardByUniqueKey(ctx context.Context, userID uint, uniqueKey string) (*models.UserCard, error)
	DeleteCard(ctx context.Context, cardID uint, userID uint) error
}

func NewUserCardRepository(db *gorm.DB) UserCardRepositoryInterface {
	return &userCardRepository{db: db}
}

func (r *userCardRepository) Create(ctx context.Context, card *models.UserCard) error {
	return r.db.WithContext(ctx).Create(card).Error
}

func (r *userCardRepository) GetCardByUserID(ctx context.Context, userID uint) ([]models.UserCard, error) {
	var cards []models.UserCard
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&cards).Error
	return cards, err
}

func (r *userCardRepository) CheckCardExists(ctx context.Context, userID uint, paymentMethodID string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&models.UserCard{}).
		Where("user_id = ? AND stripe_payment_method_id = ?", userID, paymentMethodID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *userCardRepository) GetCardByID(ctx context.Context, userID uint, cardID uint) (*models.UserCard, error) {
	var card models.UserCard

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, cardID).
		First(&card).Error; err != nil {

		return nil, err
	}

	return &card, nil
}

func (r *userCardRepository) GetCardByUniqueKey(ctx context.Context, userID uint, uniqueKey string) (*models.UserCard, error) {
	var card models.UserCard
	err := r.db.WithContext(ctx).Where("user_id = ? AND card_unique_key = ?", userID, uniqueKey).First(&card).Error
	return &card, err
}

func (r *userCardRepository) GetCardByStripePaymentMethodID(ctx context.Context, userID uint, stripePaymentMethodID string) (*models.UserCard, error) {
	var card models.UserCard

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND stripe_payment_method_id = ?", userID, stripePaymentMethodID).
		First(&card).Error; err != nil {

		return nil, err
	}

	return &card, nil
}

func (r *userCardRepository) DeleteCard(ctx context.Context, cardID uint, userID uint) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", cardID, userID).Delete(&models.UserCard{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no card found or unauthorized")
	}

	return nil
}
