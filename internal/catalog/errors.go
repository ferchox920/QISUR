package catalog

import "errors"

var (
	ErrNotImplemented          = errors.New("not implemented")
	ErrRepositoryNotConfigured = errors.New("repository not configured")
	ErrInvalidCategory         = errors.New("invalid category")
	ErrInvalidCategoryID       = errors.New("invalid category id")
	ErrInvalidProduct          = errors.New("invalid product")
	ErrInvalidProductID        = errors.New("invalid product id")
)
