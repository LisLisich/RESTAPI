package identity_jwt_provider

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
)

const privateKeyEnvironment = "JWT_ED25519_PRIVATE_KEY"

type Config struct {
	PrivateKey ed25519.PrivateKey
}

func NewConfig() (Config, error) {
	encodedKey := os.Getenv(privateKeyEnvironment)
	if encodedKey == "" {
		return Config{}, fmt.Errorf("%s is required", privateKeyEnvironment)
	}
	keyBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return Config{}, fmt.Errorf("decode %s: %w", privateKeyEnvironment, err)
	}
	if len(keyBytes) != ed25519.PrivateKeySize {
		return Config{}, fmt.Errorf(
			"%s must contain %d decoded bytes",
			privateKeyEnvironment,
			ed25519.PrivateKeySize,
		)
	}
	return Config{PrivateKey: ed25519.PrivateKey(keyBytes)}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get JWT config: %w", err))
	}
	return config
}
