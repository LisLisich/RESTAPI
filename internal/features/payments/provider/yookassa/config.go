package yookassa

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const defaultAPIURL = "https://api.yookassa.ru"

type Config struct {
	Enabled   bool
	APIURL    string
	ShopID    string
	SecretKey string
	ReturnURL string
}

func NewConfig() (Config, error) {
	enabled, err := strconv.ParseBool(environmentOrDefault("YOOKASSA_ENABLED", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse YOOKASSA_ENABLED: %w", err)
	}
	config := Config{
		Enabled:   enabled,
		APIURL:    environmentOrDefault("YOOKASSA_API_URL", defaultAPIURL),
		ShopID:    strings.TrimSpace(os.Getenv("YOOKASSA_SHOP_ID")),
		SecretKey: os.Getenv("YOOKASSA_SECRET_KEY"),
		ReturnURL: strings.TrimSpace(os.Getenv("YOOKASSA_RETURN_URL")),
	}
	if !config.Enabled {
		return config, nil
	}
	if config.ShopID == "" || config.SecretKey == "" || config.ReturnURL == "" {
		return Config{}, fmt.Errorf(
			"YOOKASSA_SHOP_ID, YOOKASSA_SECRET_KEY and YOOKASSA_RETURN_URL are required",
		)
	}
	returnURL, err := url.Parse(config.ReturnURL)
	if err != nil || returnURL.Scheme != "https" || returnURL.Host == "" {
		return Config{}, fmt.Errorf("YOOKASSA_RETURN_URL must be an absolute HTTPS URL")
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get YooKassa config: %w", err))
	}
	return config
}

func environmentOrDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
