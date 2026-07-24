package yookassa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/LisLisich/fintask/internal/core/domain"
	payments_provider "github.com/LisLisich/fintask/internal/features/payments/provider"
)

const maxResponseSize = 1 << 20

type CreatePaymentInput = payments_provider.CreatePaymentInput
type Payment = payments_provider.Payment

type Client struct {
	baseURL    string
	shopID     string
	secretKey  string
	httpClient *http.Client
}

func NewClient(
	baseURL string,
	shopID string,
	secretKey string,
	httpClient *http.Client,
) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		shopID:     shopID,
		secretKey:  secretKey,
		httpClient: httpClient,
	}
}

func (client *Client) CreatePayment(
	ctx context.Context,
	input CreatePaymentInput,
) (Payment, error) {
	body := struct {
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Capture      bool `json:"capture"`
		Confirmation struct {
			Type      string `json:"type"`
			ReturnURL string `json:"return_url"`
		} `json:"confirmation"`
		Description string `json:"description"`
	}{}
	body.Amount.Value = formatMinorUnits(input.Amount.MinorUnits)
	body.Amount.Currency = string(input.Amount.Currency)
	body.Capture = true
	body.Confirmation.Type = "redirect"
	body.Confirmation.ReturnURL = input.ReturnURL
	body.Description = input.Description

	return client.send(
		ctx,
		http.MethodPost,
		"/v3/payments",
		input.IdempotenceKey,
		body,
	)
}

func (client *Client) GetPayment(ctx context.Context, paymentID string) (Payment, error) {
	return client.send(
		ctx,
		http.MethodGet,
		"/v3/payments/"+paymentID,
		"",
		nil,
	)
}

func (client *Client) send(
	ctx context.Context,
	method string,
	path string,
	idempotenceKey string,
	body any,
) (Payment, error) {
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return Payment{}, fmt.Errorf("encode YooKassa request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(
		ctx,
		method,
		client.baseURL+path,
		requestBody,
	)
	if err != nil {
		return Payment{}, fmt.Errorf("create YooKassa request: %w", err)
	}
	request.SetBasicAuth(client.shopID, client.secretKey)
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if idempotenceKey != "" {
		request.Header.Set("Idempotence-Key", idempotenceKey)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return Payment{}, fmt.Errorf("send YooKassa request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, maxResponseSize))
		return Payment{}, fmt.Errorf(
			"YooKassa returned %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var payload struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Paid   bool   `json:"paid"`
		Test   bool   `json:"test"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseSize)).Decode(&payload); err != nil {
		return Payment{}, fmt.Errorf("decode YooKassa response: %w", err)
	}
	minorUnits, err := parseMinorUnits(payload.Amount.Value)
	if err != nil {
		return Payment{}, fmt.Errorf("parse YooKassa amount: %w", err)
	}
	amount, err := domain.NewMoney(minorUnits, domain.Currency(payload.Amount.Currency))
	if err != nil {
		return Payment{}, fmt.Errorf("validate YooKassa amount: %w", err)
	}
	return Payment{
		ID:              payload.ID,
		Status:          payload.Status,
		Paid:            payload.Paid,
		Test:            payload.Test,
		Amount:          amount,
		ConfirmationURL: payload.Confirmation.URL,
	}, nil
}

func formatMinorUnits(value int64) string {
	whole := value / 100
	fraction := value % 100
	if fraction < 0 {
		fraction = -fraction
	}
	return fmt.Sprintf("%d.%02d", whole, fraction)
}

func parseMinorUnits(value string) (int64, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		return 0, fmt.Errorf("amount must have two decimal places")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, fmt.Errorf("invalid whole amount")
	}
	fraction, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || fraction < 0 || fraction > 99 {
		return 0, fmt.Errorf("invalid fractional amount")
	}
	if whole > (1<<63-1-fraction)/100 {
		return 0, fmt.Errorf("amount overflows int64")
	}
	return whole*100 + fraction, nil
}
