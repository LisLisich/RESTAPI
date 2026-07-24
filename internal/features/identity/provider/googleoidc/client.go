package googleoidc

import (
	"context"
	"fmt"
	"net/http"

	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Client struct {
	oauthConfig oauth2.Config
	verifier    *oidc.IDTokenVerifier
}

func NewClient(ctx context.Context, config Config, httpClient *http.Client) (*Client, error) {
	if httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	}
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		return nil, fmt.Errorf("discover Google OIDC provider: %w", err)
	}
	return &Client{
		oauthConfig: oauth2.Config{
			ClientID: config.ClientID, ClientSecret: config.ClientSecret,
			RedirectURL: config.RedirectURL, Endpoint: provider.Endpoint(),
			Scopes: []string{oidc.ScopeOpenID, "email", "profile"},
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: config.ClientID}),
	}, nil
}

func (client *Client) AuthorizationURL(
	state string,
	nonce string,
	codeChallenge string,
) string {
	return client.oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

func (client *Client) Exchange(
	ctx context.Context,
	code string,
	codeVerifier string,
	expectedNonce string,
) (identity_service.GoogleIdentity, error) {
	token, err := client.oauthConfig.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return identity_service.GoogleIdentity{}, fmt.Errorf("exchange code: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return identity_service.GoogleIdentity{}, fmt.Errorf("Google response has no ID token")
	}
	idToken, err := client.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return identity_service.GoogleIdentity{}, fmt.Errorf("verify Google ID token: %w", err)
	}
	var claims struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Nonce         string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return identity_service.GoogleIdentity{}, fmt.Errorf("decode Google ID token claims: %w", err)
	}
	if claims.Nonce != expectedNonce {
		return identity_service.GoogleIdentity{}, fmt.Errorf("Google ID token nonce mismatch")
	}
	return identity_service.GoogleIdentity{
		Subject: claims.Subject, Email: claims.Email,
		EmailVerified: claims.EmailVerified, Name: claims.Name,
	}, nil
}
