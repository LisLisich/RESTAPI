package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2idConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func DefaultArgon2idConfig() Argon2idConfig {
	return Argon2idConfig{
		Memory:      19 * 1024,
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

type Argon2idHasher struct {
	config Argon2idConfig
}

func NewArgon2idHasher(config Argon2idConfig) *Argon2idHasher {
	return &Argon2idHasher{config: config}
}

func (h *Argon2idHasher) Hash(password string) (string, error) {
	if err := validateConfig(h.config); err != nil {
		return "", err
	}

	salt := make([]byte, h.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	key := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Iterations,
		h.config.Memory,
		h.config.Parallelism,
		h.config.KeyLength,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.config.Memory,
		h.config.Iterations,
		h.config.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (h *Argon2idHasher) Verify(password string, encodedHash string) (bool, error) {
	config, salt, expectedKey, err := parseEncodedHash(encodedHash)
	if err != nil {
		return false, err
	}

	actualKey := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		uint32(len(expectedKey)),
	)

	return subtle.ConstantTimeCompare(actualKey, expectedKey) == 1, nil
}

func parseEncodedHash(encodedHash string) (Argon2idConfig, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Argon2idConfig{}, nil, nil, fmt.Errorf("invalid argon2id hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return Argon2idConfig{}, nil, nil, fmt.Errorf("invalid argon2id version")
	}

	var config Argon2idConfig
	if _, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&config.Memory,
		&config.Iterations,
		&config.Parallelism,
	); err != nil {
		return Argon2idConfig{}, nil, nil, fmt.Errorf("invalid argon2id parameters")
	}
	if err := validateConfig(config); err != nil {
		return Argon2idConfig{}, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 16 {
		return Argon2idConfig{}, nil, nil, fmt.Errorf("invalid argon2id salt")
	}

	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expectedKey) < 16 || len(expectedKey) > 64 {
		return Argon2idConfig{}, nil, nil, fmt.Errorf("invalid argon2id key")
	}

	return config, salt, expectedKey, nil
}

func validateConfig(config Argon2idConfig) error {
	if config.Memory < 19*1024 || config.Memory > 64*1024 {
		return fmt.Errorf("argon2id memory must be between 19456 and 65536 KiB")
	}
	if config.Iterations < 2 || config.Iterations > 10 {
		return fmt.Errorf("argon2id iterations must be between 2 and 10")
	}
	if config.Parallelism < 1 || config.Parallelism > 8 {
		return fmt.Errorf("argon2id parallelism must be between 1 and 8")
	}
	if config.SaltLength != 0 && (config.SaltLength < 16 || config.SaltLength > 64) {
		return fmt.Errorf("argon2id salt length must be between 16 and 64 bytes")
	}
	if config.KeyLength != 0 && (config.KeyLength < 16 || config.KeyLength > 64) {
		return fmt.Errorf("argon2id key length must be between 16 and 64 bytes")
	}
	return nil
}
