package utils

import (
	"time"

	"github.com/NOOKX2/e-commerce-backend/internal/domain"
	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(user *domain.User, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":     user.Email,
		"name":      user.Name,
		"user_id":   user.ID,
		"user_role": user.Role,
		"exp":       time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString([]byte(secretKey))
}
