package identity_jwt_provider

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type Claims struct {
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	Subject   string `json:"sub"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	JWTID     string `json:"jti"`
	TokenType string `json:"typ"`
}

type Ed25519Provider struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	audience   string
	ttl        time.Duration
	now        func() time.Time
}

func NewEd25519Provider(
	privateKey ed25519.PrivateKey,
	publicKey ed25519.PublicKey,
	issuer string,
	audience string,
	ttl time.Duration,
	now func() time.Time,
) *Ed25519Provider {
	return &Ed25519Provider{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		audience:   audience,
		ttl:        ttl,
		now:        now,
	}
}

func (provider *Ed25519Provider) Issue(userID int, jwtID string) (string, error) {
	if len(provider.privateKey) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("invalid Ed25519 private key")
	}
	if userID <= 0 || strings.TrimSpace(jwtID) == "" {
		return "", fmt.Errorf("invalid access token identity")
	}
	now := provider.now().UTC()
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "EdDSA",
		"typ": "JWT",
	})
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}
	claimsJSON, err := json.Marshal(Claims{
		Issuer:    provider.issuer,
		Audience:  provider.audience,
		Subject:   strconv.Itoa(userID),
		ExpiresAt: now.Add(provider.ttl).Unix(),
		IssuedAt:  now.Unix(),
		JWTID:     jwtID,
		TokenType: "access",
	})
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := header + "." + payload
	signature := ed25519.Sign(provider.privateKey, []byte(signingInput))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (provider *Ed25519Provider) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, unauthorizedTokenError("invalid token segments")
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, unauthorizedTokenError("invalid token header encoding")
	}
	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil ||
		header.Algorithm != "EdDSA" ||
		header.Type != "JWT" {
		return Claims{}, unauthorizedTokenError("invalid token header")
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(provider.publicKey) != ed25519.PublicKeySize {
		return Claims{}, unauthorizedTokenError("invalid token signature encoding")
	}
	signingInput := parts[0] + "." + parts[1]
	if !ed25519.Verify(provider.publicKey, []byte(signingInput), signature) {
		return Claims{}, unauthorizedTokenError("invalid token signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, unauthorizedTokenError("invalid claims encoding")
	}
	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return Claims{}, unauthorizedTokenError("invalid claims")
	}
	now := provider.now().UTC().Unix()
	userID, subjectErr := strconv.Atoi(claims.Subject)
	if claims.Issuer != provider.issuer ||
		claims.Audience != provider.audience ||
		claims.TokenType != "access" ||
		claims.JWTID == "" ||
		subjectErr != nil ||
		userID <= 0 ||
		claims.ExpiresAt <= now ||
		claims.IssuedAt > now+60 {
		return Claims{}, unauthorizedTokenError("invalid registered claims")
	}
	return claims, nil
}

func unauthorizedTokenError(reason string) error {
	return fmt.Errorf("%s: %w", reason, core_errors.ErrUnauthorized)
}
