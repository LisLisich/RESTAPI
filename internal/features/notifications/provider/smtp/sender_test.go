package smtp

import (
	"strings"
	"testing"
)

func TestBuildMessageUsesUTF8Headers(t *testing.T) {
	message, err := buildMessage(
		"noreply@example.com",
		"user@example.com",
		"Кошелек пополнен",
		"На кошелек зачислено 125.50 RUB.",
	)

	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	text := string(message)
	if !strings.Contains(text, "Content-Type: text/plain; charset=UTF-8") ||
		!strings.Contains(text, "На кошелек зачислено") {
		t.Fatalf("unexpected message: %s", text)
	}
}

func TestBuildMessageRejectsHeaderInjection(t *testing.T) {
	_, err := buildMessage(
		"noreply@example.com\r\nBcc: attacker@example.com",
		"user@example.com",
		"subject",
		"body",
	)

	if err == nil {
		t.Fatal("expected header injection error")
	}
}
