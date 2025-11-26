package identity

import "context"

// PasswordHasher abstrae la estrategia de hashing (bcrypt, argon2, etc.).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// TokenProvider emite tokens de auth para usuarios autenticados.
type TokenProvider interface {
	Generate(ctx context.Context, user User) (string, error)
}

// VerificationCodeGenerator genera codigos OTP para verificacion por email.
type VerificationCodeGenerator interface {
	Generate(ctx context.Context, userID string) (string, error)
}

// VerificationSender envia desafios de verificacion (email, SMS, etc.).
type VerificationSender interface {
	SendVerification(ctx context.Context, email, code string) error
}
