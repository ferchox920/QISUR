package catalog

import "errors"

var (
	ErrNotImplemented          = errors.New("not implemented")
	ErrRepositoryNotConfigured = errors.New("repository not configured")
)
