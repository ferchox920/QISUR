package identity

import "context"

// PasswordHasher abstracts hashing strategy (e.g., bcrypt, argon2).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// TokenProvider issues auth tokens for signed-in users.
type TokenProvider interface {
	Generate(ctx context.Context, user User) (string, error)
}

// VerificationCodeGenerator issues OTP codes for email verification.
type VerificationCodeGenerator interface {
	Generate(ctx context.Context, userID string) (string, error)
}

// VerificationSender sends verification challenges (email, SMS, etc.).
type VerificationSender interface {
	SendVerification(ctx context.Context, email, code string) error
}
