package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunLoginReadsPasswordFromEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var request map[string]string
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request["email"] != "user@example.com" || request["password"] != "secret" {
			t.Fatalf("unexpected request: %#v", request)
		}
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{
			"access_token":"access",
			"refresh_token":"refresh",
			"token_type":"Bearer",
			"expires_in":900
		}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{"login", "--api-url", server.URL, "--email", "user@example.com"},
		&stdout,
		&stderr,
		func(name string) string {
			if name == "FINTASK_PASSWORD" {
				return "secret"
			}
			return ""
		},
	)

	if exitCode != 0 {
		t.Fatalf("expected success, got %d: %s", exitCode, stderr.String())
	}
	var response map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if response["access_token"] != "access" || response["refresh_token"] != "refresh" {
		t.Fatalf("unexpected output: %#v", response)
	}
}

func TestRunWalletReadsAccessTokenFromEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer access" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{
			"version":2,
			"balance_minor":15000,
			"currency":"RUB"
		}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{"wallet", "--api-url", server.URL},
		&stdout,
		&stderr,
		func(name string) string {
			if name == "FINTASK_ACCESS_TOKEN" {
				return "access"
			}
			return ""
		},
	)

	if exitCode != 0 {
		t.Fatalf("expected success, got %d: %s", exitCode, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"balance_minor": 15000`)) {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestRunKeygenProducesEd25519PrivateKey(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"keygen"}, &stdout, &stderr, func(string) string { return "" })

	if exitCode != 0 {
		t.Fatalf("expected success, got %d: %s", exitCode, stderr.String())
	}
	decoded, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(stdout.Bytes())))
	if err != nil {
		t.Fatalf("decode key: %v", err)
	}
	if len(decoded) != ed25519.PrivateKeySize {
		t.Fatalf("expected %d bytes, got %d", ed25519.PrivateKeySize, len(decoded))
	}
}

func TestRunRejectsMissingSecret(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{"login", "--email", "user@example.com"},
		&stdout,
		&stderr,
		func(string) string { return "" },
	)

	if exitCode != 2 {
		t.Fatalf("expected usage error, got %d", exitCode)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("FINTASK_PASSWORD")) {
		t.Fatalf("expected environment variable hint, got %s", stderr.String())
	}
}
