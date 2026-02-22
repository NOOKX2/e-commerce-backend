package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/response"
	"github.com/minio/minio-go/v7"
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
		objectName := extractObjectNameFromURL(existingMedia.PublicURL, s.r2.PublicURL)

		_, errR2 := s.r2.Client.StatObject(context.Background(), s.r2.BucketName, objectName, minio.StatObjectOptions{})

		if errR2 == nil {
			return &response.UploadInstructionResponse{
				UploadURL: "",
				PublicURL: existingMedia.PublicURL,
				Exists:    true,
			}, nil
		}
		fmt.Println("Warning: Found DB record but missing in R2. Cleaning up DB...")

		if err := s.uploadRepo.DeleteMediaByHash(hash); err != nil {
			fmt.Printf("Error: Failed to delete ghost record from DB: %v\n", err)
		} else {
			fmt.Println("Hash delete successfully")
		}
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

func (s *uploadService) ConfirmUpload(hash, publicURL string) error {
    return s.SaveMediaRecord(hash, publicURL)
}

func extractObjectNameFromURL(fullURL, baseURL string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	objectName := strings.TrimPrefix(fullURL, baseURL)
	objectName = strings.TrimPrefix(objectName, "/")
	return objectName
}