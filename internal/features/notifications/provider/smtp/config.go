package smtp

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Enabled  bool
	Address  string
	Host     string
	Username string
	Password string
	From     string
}

func NewConfig() (Config, error) {
	enabled, err := strconv.ParseBool(environmentOrDefault("SMTP_ENABLED", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SMTP_ENABLED: %w", err)
	}
	config := Config{
		Enabled:  enabled,
		Host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
	}
	port := environmentOrDefault("SMTP_PORT", "587")
	config.Address = net.JoinHostPort(config.Host, port)
	if enabled &&
		(config.Host == "" || config.Username == "" || config.Password == "" || config.From == "") {
		return Config{}, fmt.Errorf(
			"SMTP_HOST, SMTP_USERNAME, SMTP_PASSWORD and SMTP_FROM are required",
		)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get SMTP config: %w", err))
	}
	return config
}

func environmentOrDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
