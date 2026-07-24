package googleoidc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Enabled      bool
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewConfig() (Config, error) {
	enabled, err := strconv.ParseBool(environmentOrDefault("GOOGLE_OIDC_ENABLED", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse GOOGLE_OIDC_ENABLED: %w", err)
	}
	config := Config{
		Enabled: enabled, Issuer: environmentOrDefault(
			"GOOGLE_OIDC_ISSUER",
			"https://accounts.google.com",
		),
		ClientID:     strings.TrimSpace(os.Getenv("GOOGLE_OIDC_CLIENT_ID")),
		ClientSecret: os.Getenv("GOOGLE_OIDC_CLIENT_SECRET"),
		RedirectURL:  strings.TrimSpace(os.Getenv("GOOGLE_OIDC_REDIRECT_URL")),
	}
	if enabled &&
		(config.ClientID == "" || config.ClientSecret == "" || config.RedirectURL == "") {
		return Config{}, fmt.Errorf(
			"GOOGLE_OIDC_CLIENT_ID, GOOGLE_OIDC_CLIENT_SECRET and GOOGLE_OIDC_REDIRECT_URL are required",
		)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get Google OIDC config: %w", err))
	}
	return config
}

func environmentOrDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
