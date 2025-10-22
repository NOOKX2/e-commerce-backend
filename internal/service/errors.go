package service

import "errors"

var (
	ErrProductNotFound = errors.New("product not found")
	ErrForbidden       = errors.New("forbidden: you are not the owner of this product")
	ErrFailedToMapData = errors.New("failed to map update data")
)
