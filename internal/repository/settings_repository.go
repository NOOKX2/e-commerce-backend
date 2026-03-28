package repository

import (
	"context"
	"errors"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type SettingsRepositoryInterface interface {
	GetOrCreatePlatformSettings(ctx context.Context) (*models.PlatformSetting, error)
	UpdatePlatformSettings(ctx context.Context, s *models.PlatformSetting) error
	IsManualProductApprovalEnabled(ctx context.Context) (bool, error)

	GetOrCreateSellerShop(ctx context.Context, sellerID uint) (*models.SellerShop, error)
	UpsertSellerShop(ctx context.Context, shop *models.SellerShop) error
}

type settingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) SettingsRepositoryInterface {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) GetOrCreatePlatformSettings(ctx context.Context) (*models.PlatformSetting, error) {
	var s models.PlatformSetting
	err := r.db.WithContext(ctx).First(&s, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s = models.PlatformSetting{}
		s.ID = 1
		s.SiteName = "E-Commerce"
		s.Currency = "THB"
		if err := r.db.WithContext(ctx).Create(&s).Error; err != nil {
			return nil, err
		}
		return &s, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settingsRepository) UpdatePlatformSettings(ctx context.Context, s *models.PlatformSetting) error {
	s.ID = 1
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *settingsRepository) IsManualProductApprovalEnabled(ctx context.Context) (bool, error) {
	s, err := r.GetOrCreatePlatformSettings(ctx)
	if err != nil {
		return false, err
	}
	return s.ManualProductApproval, nil
}

func (r *settingsRepository) GetOrCreateSellerShop(ctx context.Context, sellerID uint) (*models.SellerShop, error) {
	var shop models.SellerShop
	err := r.db.WithContext(ctx).Where("seller_id = ?", sellerID).First(&shop).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		shop = models.SellerShop{SellerID: sellerID}
		if err := r.db.WithContext(ctx).Create(&shop).Error; err != nil {
			return nil, err
		}
		return &shop, nil
	}
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *settingsRepository) UpsertSellerShop(ctx context.Context, shop *models.SellerShop) error {
	var existing models.SellerShop
	err := r.db.WithContext(ctx).Where("seller_id = ?", shop.SellerID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(shop).Error
	}
	if err != nil {
		return err
	}
	shop.ID = existing.ID
	shop.CreatedAt = existing.CreatedAt
	return r.db.WithContext(ctx).Save(shop).Error
}
