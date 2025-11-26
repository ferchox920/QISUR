package http

import (
	"catalog-api/pkg/crypto"
)

// JWTValidatorAdapter conecta JWTProvider con el middleware TokenValidator.
type JWTValidatorAdapter struct {
	Provider crypto.JWTProvider
}

func (j JWTValidatorAdapter) Validate(token string) (AuthContext, error) {
	claims, err := j.Provider.Validate(token)
	if err != nil {
		return AuthContext{}, err
	}
	return AuthContext{
		UserID: claims.Subject,
		Role:   claims.Role,
	}, nil
}
