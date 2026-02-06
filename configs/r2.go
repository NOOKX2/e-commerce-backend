package configs

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type R2Config struct {
	Client     *minio.Client
	BucketName string
	PublicURL  string
}

func LoadR2Config() *R2Config {
	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("ACCESS_KEY_ID")
	secretKey := os.Getenv("SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET_NAME")
	publicUrl := os.Getenv("R2_PUBLIC_URL")

	if endpoint == "" || accessKey == "" || secretKey == "" {
		log.Fatal("R2 configuration is missing in .env file")
	}

	finalEndpoint := endpoint
	if strings.Contains(endpoint, "://") {
		u, err := url.Parse(endpoint)
		if err == nil {
			finalEndpoint = u.Host 
		}
	}

	minioClient, err := minio.New(finalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})

	if err != nil {
		log.Fatalln("Error creating R2 client:", err)
	}

	log.Println("Successfully connected to Cloudflare R2")

	return &R2Config{
		Client:     minioClient,
		BucketName: bucket,
		PublicURL:  publicUrl,
	}
}
