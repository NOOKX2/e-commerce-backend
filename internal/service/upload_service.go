package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/response"
	"gorm.io/gorm"
)

type UploadService interface {
	GetUploadInstructions(hash, filename, contentType string) (*response.UploadInstructionResponse, error)
	SaveMediaRecord(hash, publicURL string) error
}

type uploadService struct {
	r2         *configs.R2Config
	uploadRepo repository.UploadRepository
}

func NewUploadService(r2 *configs.R2Config, repo repository.UploadRepository) UploadService {
	return &uploadService{r2: r2, uploadRepo: repo}
}

func (s *uploadService) GetUploadInstructions(hash, filename, contentType string) (*response.UploadInstructionResponse, error) {
	existingMedia, err := s.uploadRepo.GetMediaByHash(hash)
	if err == nil && existingMedia != nil {
		return &response.UploadInstructionResponse{
			UploadURL: "",
			PublicURL: existingMedia.PublicURL,
			Exists:    true,
		}, nil
	}

	objectName := fmt.Sprintf("products/%d-%s", time.Now().Unix(), filename)
	presignedURL, err := s.r2.Client.PresignedPutObject(
		context.Background(),
		s.r2.BucketName,
		objectName,
		time.Minute*15,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed URL: %v", err)
	}

	return &response.UploadInstructionResponse{
		UploadURL: presignedURL.String(),
		PublicURL: fmt.Sprintf("%s/%s", s.r2.PublicURL, objectName),
		Exists:    false,
	}, nil
}

func (s *uploadService) SaveMediaRecord(hash, publicURL string) error {
	existing, err := s.uploadRepo.GetMediaByHash(hash)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existing != nil {
		return nil
	}

	media := &models.Media{
		Hash:      hash,
		PublicURL: publicURL,
	}
	
	return s.uploadRepo.SaveMedia(media)
}
