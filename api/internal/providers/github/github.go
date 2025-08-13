package github

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Verifier struct {
	appID      int64
	privateKey *rsa.PrivateKey
	http       *http.Client
}

// NewVerifierFromEnv loads APP_ID + PRIVATE_KEY (PEM) from env.
func NewVerifierFromEnv() (*Verifier, error) {
	idStr := os.Getenv("GITHUB_APP_ID")
	if idStr == "" {
		return nil, errors.New("GITHUB_APP_ID not set")
	}
	appID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid GITHUB_APP_ID: %w", err)
	}

	key := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if key == "" {
		return nil, errors.New("GITHUB_APP_PRIVATE_KEY not set")
	}
	// support "\n" in env vars
	key = strings.ReplaceAll(key, `\n`, "\n")
	priv, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	return &Verifier{
		appID:      appID,
		privateKey: priv,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (v *Verifier) signAppJWT() (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iat": now.Add(-1 * time.Minute).Unix(), // backdate to avoid drift
		"exp": now.Add(9 * time.Minute).Unix(),  // max 10 min
		"iss": v.appID,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tok.SignedString(v.privateKey)
}

func (v *Verifier) VerifyAccess(ctx context.Context, owner, repo string) error {
	if owner == "" || repo == "" {
		return errors.New("owner and repo are required")
	}
	appJWT, err := v.signAppJWT()
	if err != nil {
		return fmt.Errorf("github app jwt: %w", err)
	}

	// 1) Find installation for this repository
	installURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/installation", owner, repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", installURL, nil)
	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := v.http.Do(req)
	if err != nil {
		return fmt.Errorf("github installation lookup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("GitHub App is not installed on %s/%s", owner, repo)
	}
	if resp.StatusCode != 200 {
		var body struct{ Message string }
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return fmt.Errorf("installation lookup failed (%d): %s", resp.StatusCode, body.Message)
	}
	var inst struct{ ID int64 `json:"id"` }
	if err := json.NewDecoder(resp.Body).Decode(&inst); err != nil {
		return fmt.Errorf("decode installation: %w", err)
	}
	if inst.ID == 0 {
		return errors.New("installation id missing")
	}

	// 2) Mint an installation access token
	tokURL := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", inst.ID)
	req2, _ := http.NewRequestWithContext(ctx, "POST", tokURL, nil)
	req2.Header.Set("Authorization", "Bearer "+appJWT)
	req2.Header.Set("Accept", "application/vnd.github+json")

	resp2, err := v.http.Do(req2)
	if err != nil {
		return fmt.Errorf("create installation token: %w", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 201 {
		var body struct{ Message string }
		_ = json.NewDecoder(resp2.Body).Decode(&body)
		return fmt.Errorf("installation token failed (%d): %s", resp2.StatusCode, body.Message)
	}
	var tok struct{ Token string `json:"token"` }
	if err := json.NewDecoder(resp2.Body).Decode(&tok); err != nil {
		return fmt.Errorf("decode token: %w", err)
	}
	if tok.Token == "" {
		return errors.New("empty installation token")
	}

	// 3) Probe a trivial endpoint using the installation token to prove access
	repoURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req3, _ := http.NewRequestWithContext(ctx, "GET", repoURL, nil)
	req3.Header.Set("Authorization", "Bearer "+tok.Token)
	req3.Header.Set("Accept", "application/vnd.github+json")

	resp3, err := v.http.Do(req3)
	if err != nil {
		return fmt.Errorf("repo probe: %w", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != 200 {
		var body struct{ Message string }
		_ = json.NewDecoder(resp3.Body).Decode(&body)
		return fmt.Errorf("repo access denied (%d): %s", resp3.StatusCode, body.Message)
	}

	return nil
}

// InstallationTokenForRepo returns an installation token for the app on owner/repo.
func (v *Verifier) InstallationTokenForRepo(ctx context.Context, owner, repo string) (string, error) {
	if owner == "" || repo == "" {
		return "", errors.New("owner and repo are required")
	}
	appJWT, err := v.signAppJWT()
	if err != nil {
		return "", fmt.Errorf("github app jwt: %w", err)
	}

	// 1) Find installation for this repository
	installURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/installation", owner, repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", installURL, nil)
	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := v.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("github installation lookup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("GitHub App is not installed on %s/%s", owner, repo)
	}
	if resp.StatusCode != 200 {
		var body struct{ Message string }
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return "", fmt.Errorf("installation lookup failed (%d): %s", resp.StatusCode, body.Message)
	}
	var inst struct{ ID int64 `json:"id"` }
	if err := json.NewDecoder(resp.Body).Decode(&inst); err != nil {
		return "", fmt.Errorf("decode installation: %w", err)
	}
	if inst.ID == 0 {
	 return "", errors.New("installation id missing")
	}

	// 2) Mint an installation access token
	tokURL := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", inst.ID)
	req2, _ := http.NewRequestWithContext(ctx, "POST", tokURL, nil)
	req2.Header.Set("Authorization", "Bearer "+appJWT)
	req2.Header.Set("Accept", "application/vnd.github+json")

	resp2, err := v.http.Do(req2)
	if err != nil {
		return "", fmt.Errorf("create installation token: %w", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 201 {
		var body struct{ Message string }
		_ = json.NewDecoder(resp2.Body).Decode(&body)
		return "", fmt.Errorf("installation token failed (%d): %s", resp2.StatusCode, body.Message)
	}
	var tok struct{ Token string `json:"token"` }
	if err := json.NewDecoder(resp2.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	if tok.Token == "" {
		return "", errors.New("empty installation token")
	}
	return tok.Token, nil
}
