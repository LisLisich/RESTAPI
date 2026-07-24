package identity_token_provider

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
)

const minimumTokenLength = 16

type RandomIssuer struct {
	tokenLength int
}

func NewRandomIssuer(tokenLength int) *RandomIssuer {
	return &RandomIssuer{tokenLength: tokenLength}
}

func (i *RandomIssuer) Issue() (identity_service.IssuedToken, error) {
	if i.tokenLength < minimumTokenLength {
		return identity_service.IssuedToken{}, fmt.Errorf(
			"token length must be at least %d bytes",
			minimumTokenLength,
		)
	}

	randomBytes := make([]byte, i.tokenLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return identity_service.IssuedToken{}, fmt.Errorf("read secure random bytes: %w", err)
	}

	rawToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	return identity_service.IssuedToken{
		Raw:  rawToken,
		Hash: i.Hash(rawToken),
	}, nil
}

func (*RandomIssuer) Hash(rawToken string) []byte {
	tokenHash := sha256.Sum256([]byte(rawToken))
	return tokenHash[:]
}
