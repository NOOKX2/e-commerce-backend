package request

import "github.com/NOOKX2/e-commerce-backend/internal/models"

type UpdateUserStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type UpdateProductStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type UpdateUserDetailsRequest struct {
	Name   string            `json:"name" validate:"required"`
	Role   models.UserRole   `json:"role" validate:"required,oneof=buyer seller admin"`
	Status models.UserStatus `json:"status" validate:"required,oneof=active suspended banned"`
}

