package service

import "errors"

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrForbidden         = errors.New("forbidden: you are not the owner of this product")
	ErrFailedToMapData   = errors.New("failed to map update data")
	ErrUserExisted       = errors.New("user email aleady existed")
	ErrUserNotFound      = errors.New("user email not found")
	ErrPasswordIncorrect = errors.New("password incorrect")
	ErrAccountSuspended  = errors.New("account suspended")
	ErrAccountBanned     = errors.New("account banned")
	ErrOrderNotFound     = errors.New("order not found")
)
