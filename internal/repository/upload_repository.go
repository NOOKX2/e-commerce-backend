package repository

import (
	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type UploadRepository interface {
	GetMediaByHash(hash string) (*models.Media, error)
	SaveMedia(media *models.Media) error
}

type uploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) UploadRepository {
    return &uploadRepository{db: db}
}

func (r *uploadRepository) GetMediaByHash(hash string) (*models.Media, error) {
	var media models.Media

	result := r.db.Where("hash = ?", hash).Limit(1).Find(&media)
    if result.RowsAffected == 0 {
        return nil, gorm.ErrRecordNotFound
    }

	return &media, nil
}

func (r *uploadRepository) SaveMedia(media *models.Media) error {
    return r.db.Create(media).Error
}
