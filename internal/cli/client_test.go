package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/auth/token" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.Email != "user@example.com" || request.Password != "secret" {
			t.Fatalf("unexpected credentials: %#v", request)
		}
		writeJSON(t, rw, http.StatusOK, TokenPair{
			AccessToken:  "access",
			RefreshToken: "refresh",
			TokenType:    "Bearer",
			ExpiresIn:    900,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	pair, err := client.Login(t.Context(), "user@example.com", "secret")

	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if pair.AccessToken != "access" || pair.RefreshToken != "refresh" {
		t.Fatalf("unexpected token pair: %#v", pair)
	}
}

func TestClientRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/auth/token/refresh" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var request struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.RefreshToken != "old-refresh" {
			t.Fatalf("unexpected refresh token %q", request.RefreshToken)
		}
		writeJSON(t, rw, http.StatusOK, TokenPair{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
			TokenType:    "Bearer",
			ExpiresIn:    900,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	pair, err := client.Refresh(t.Context(), "old-refresh")

	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if pair.AccessToken != "new-access" || pair.RefreshToken != "new-refresh" {
		t.Fatalf("unexpected token pair: %#v", pair)
	}
}

func TestClientWalletUsesBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/wallet" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		writeJSON(t, rw, http.StatusOK, Wallet{
			Version:      3,
			BalanceMinor: 12550,
			Currency:     "RUB",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	wallet, err := client.GetWallet(t.Context(), "access-token")

	if err != nil {
		t.Fatalf("get wallet: %v", err)
	}
	if wallet.BalanceMinor != 12550 || wallet.Currency != "RUB" || wallet.Version != 3 {
		t.Fatalf("unexpected wallet: %#v", wallet)
	}
}

func TestClientReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		writeJSON(t, rw, http.StatusUnauthorized, map[string]string{
			"error":   "invalid credentials",
			"message": "failed to issue API tokens",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	_, err := client.Login(t.Context(), "user@example.com", "wrong")

	if err == nil || err.Error() != "API returned 401: failed to issue API tokens" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeJSON(t *testing.T, rw http.ResponseWriter, status int, value any) {
	t.Helper()
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	if err := json.NewEncoder(rw).Encode(value); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}
