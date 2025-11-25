package identity

import "errors"

var (
	ErrEmailAlreadyRegistered   = errors.New("email already registered")
	ErrUserNotFound             = errors.New("user not found")
	ErrUserBlocked              = errors.New("user is blocked")
	ErrUserNotVerified          = errors.New("user not verified")
	ErrInvalidVerificationCode  = errors.New("invalid verification code")
	ErrRepositoryNotConfigured  = errors.New("repository not configured")
	ErrPasswordHasherNotSet     = errors.New("password hasher not configured")
	ErrVerificationSenderNotSet = errors.New("verification sender not configured")
	ErrNotImplemented           = errors.New("not implemented")
)
