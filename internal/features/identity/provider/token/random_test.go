package identity_token_provider

import (
	"bytes"
	"testing"
)

func TestRandomIssuerCreatesOpaqueUniqueTokens(t *testing.T) {
	issuer := NewRandomIssuer(32)

	first, err := issuer.Issue()
	if err != nil {
		t.Fatalf("issue first token: %v", err)
	}
	second, err := issuer.Issue()
	if err != nil {
		t.Fatalf("issue second token: %v", err)
	}

	if first.Raw == "" || second.Raw == "" {
		t.Fatal("expected non-empty raw tokens")
	}
	if first.Raw == second.Raw {
		t.Fatal("expected unique raw tokens")
	}
	if len(first.Hash) != 32 || len(second.Hash) != 32 {
		t.Fatalf("expected SHA-256 hashes, got %d and %d bytes", len(first.Hash), len(second.Hash))
	}
	if bytes.Equal(first.Hash, second.Hash) {
		t.Fatal("expected unique token hashes")
	}
	if bytes.Contains(first.Hash, []byte(first.Raw)) {
		t.Fatal("stored hash must not contain the raw token")
	}
}

func TestRandomIssuerRejectsWeakTokenLength(t *testing.T) {
	issuer := NewRandomIssuer(15)

	if _, err := issuer.Issue(); err == nil {
		t.Fatal("expected weak token length to be rejected")
	}
}
