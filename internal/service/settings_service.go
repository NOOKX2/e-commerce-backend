package service

import (
	"context"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
)

type SettingsServiceInterface interface {
	GetPlatformSettings(ctx context.Context) (*models.PlatformSetting, error)
	UpdatePlatformSettings(ctx context.Context, req request.UpdatePlatformSettingsRequest) (*models.PlatformSetting, error)

	GetSellerShop(ctx context.Context, sellerID uint) (*models.SellerShop, error)
	UpdateSellerShop(ctx context.Context, sellerID uint, req request.UpdateSellerShopRequest) (*models.SellerShop, error)
}

type SettingsService struct {
	repo repository.SettingsRepositoryInterface
}

func NewSettingsService(repo repository.SettingsRepositoryInterface) SettingsServiceInterface {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) GetPlatformSettings(ctx context.Context) (*models.PlatformSetting, error) {
	return s.repo.GetOrCreatePlatformSettings(ctx)
}

func (s *SettingsService) UpdatePlatformSettings(ctx context.Context, req request.UpdatePlatformSettingsRequest) (*models.PlatformSetting, error) {
	current, err := s.repo.GetOrCreatePlatformSettings(ctx)
	if err != nil {
		return nil, err
	}
	if req.MaintenanceMode != nil {
		current.MaintenanceMode = *req.MaintenanceMode
	}
	if req.SiteName != nil {
		current.SiteName = strings.TrimSpace(*req.SiteName)
	}
	if req.CommissionRate != nil {
		current.CommissionRate = *req.CommissionRate
	}
	if req.Currency != nil {
		current.Currency = strings.TrimSpace(strings.ToUpper(*req.Currency))
	}
	if req.ManualProductApproval != nil {
		current.ManualProductApproval = *req.ManualProductApproval
	}
	if err := s.repo.UpdatePlatformSettings(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *SettingsService) GetSellerShop(ctx context.Context, sellerID uint) (*models.SellerShop, error) {
	return s.repo.GetOrCreateSellerShop(ctx, sellerID)
}

func (s *SettingsService) UpdateSellerShop(ctx context.Context, sellerID uint, req request.UpdateSellerShopRequest) (*models.SellerShop, error) {
	shop, err := s.repo.GetOrCreateSellerShop(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	if req.ShopName != nil {
		shop.ShopName = strings.TrimSpace(*req.ShopName)
	}
	if req.Description != nil {
		shop.Description = *req.Description
	}
	if req.LogoURL != nil {
		shop.LogoURL = strings.TrimSpace(*req.LogoURL)
	}
	if req.PickupAddress != nil {
		shop.PickupAddress = strings.TrimSpace(*req.PickupAddress)
	}
	if req.BankName != nil {
		shop.BankName = strings.TrimSpace(*req.BankName)
	}
	if req.AccountNumber != nil {
		shop.AccountNumber = strings.TrimSpace(*req.AccountNumber)
	}
	if req.AccountHolder != nil {
		shop.AccountHolder = strings.TrimSpace(*req.AccountHolder)
	}
	shop.SellerID = sellerID
	if err := s.repo.UpsertSellerShop(ctx, shop); err != nil {
		return nil, err
	}
	return shop, nil
}
