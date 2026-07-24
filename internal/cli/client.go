package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxResponseSize = 1 << 20

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type Wallet struct {
	Version      int    `json:"version"`
	BalanceMinor int64  `json:"balance_minor"`
	Currency     string `json:"currency"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (client *Client) Login(
	ctx context.Context,
	email string,
	password string,
) (TokenPair, error) {
	return sendJSON[TokenPair](
		client,
		ctx,
		http.MethodPost,
		"/api/v2/auth/token",
		map[string]string{"email": email, "password": password},
		"",
	)
}

func (client *Client) Refresh(
	ctx context.Context,
	refreshToken string,
) (TokenPair, error) {
	return sendJSON[TokenPair](
		client,
		ctx,
		http.MethodPost,
		"/api/v2/auth/token/refresh",
		map[string]string{"refresh_token": refreshToken},
		"",
	)
}

func (client *Client) GetWallet(
	ctx context.Context,
	accessToken string,
) (Wallet, error) {
	return sendJSON[Wallet](
		client,
		ctx,
		http.MethodGet,
		"/api/v2/wallet",
		nil,
		accessToken,
	)
}

func sendJSON[T any](
	client *Client,
	ctx context.Context,
	method string,
	path string,
	body any,
	accessToken string,
) (T, error) {
	var zero T
	var requestBody io.Reader
	if body != nil {
		encodedBody, err := json.Marshal(body)
		if err != nil {
			return zero, fmt.Errorf("encode request: %w", err)
		}
		requestBody = bytes.NewReader(encodedBody)
	}
	request, err := http.NewRequestWithContext(
		ctx,
		method,
		client.baseURL+path,
		requestBody,
	)
	if err != nil {
		return zero, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return zero, fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()

	limitedBody := io.LimitReader(response.Body, maxResponseSize)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		var apiError struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(limitedBody).Decode(&apiError); err != nil {
			return zero, fmt.Errorf("API returned %d", response.StatusCode)
		}
		return zero, fmt.Errorf("API returned %d: %s", response.StatusCode, apiError.Message)
	}

	var result T
	if err := json.NewDecoder(limitedBody).Decode(&result); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}
