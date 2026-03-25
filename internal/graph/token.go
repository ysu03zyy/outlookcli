package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ysu03zyy/outlookcli/internal/config"
)

const (
	tokenEndpoint = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	graphScope    = "https://graph.microsoft.com/Mail.ReadWrite https://graph.microsoft.com/Mail.Send https://graph.microsoft.com/Calendars.ReadWrite offline_access"
)

// Credentials is the persisted token payload (subset of OAuth2 token response).
type Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

func credsPath(dir string) string {
	return filepath.Join(dir, "credentials.json")
}

// LoadCredentials reads credentials.json.
func LoadCredentials(dir string) (*Credentials, map[string]any, error) {
	p := credsPath(dir)
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, nil, fmt.Errorf("read credentials: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, fmt.Errorf("parse credentials: %w", err)
	}
	var c Credentials
	_ = json.Unmarshal(b, &c)
	return &c, raw, nil
}

// SaveCredentials writes merged credentials to disk.
func SaveCredentials(dir string, raw map[string]any) error {
	b, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	p := credsPath(dir)
	return os.WriteFile(p, b, 0o600)
}

// Refresh exchanges refresh_token for new tokens and updates credentials.json.
func Refresh(ctx context.Context, dir string, app *config.AppConfig) (*Credentials, error) {
	creds, raw, err := LoadCredentials(dir)
	if err != nil {
		return nil, err
	}
	if creds.RefreshToken == "" {
		return nil, fmt.Errorf("credentials.json: missing refresh_token; run Outlook setup (see README)")
	}

	form := url.Values{}
	form.Set("client_id", app.ClientID)
	form.Set("client_secret", app.ClientSecret)
	form.Set("refresh_token", creds.RefreshToken)
	form.Set("grant_type", "refresh_token")
	form.Set("scope", graphScope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token refresh failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var tok map[string]any
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, err
	}
	// Merge into existing raw so we keep refresh_token if absent in response.
	for k, v := range tok {
		raw[k] = v
	}
	if _, ok := tok["refresh_token"]; !ok {
		raw["refresh_token"] = creds.RefreshToken
	}
	if err := SaveCredentials(dir, raw); err != nil {
		return nil, err
	}

	var out Credentials
	b2, _ := json.Marshal(raw)
	_ = json.Unmarshal(b2, &out)
	return &out, nil
}

// EnsureAccessToken refreshes (always, matching shell scripts) and returns a bearer token.
func EnsureAccessToken(ctx context.Context, dir string) (string, error) {
	app, err := config.LoadAppConfig(dir)
	if err != nil {
		return "", err
	}
	creds, err := Refresh(ctx, dir, app)
	if err != nil {
		return "", err
	}
	if creds.AccessToken == "" {
		return "", fmt.Errorf("no access_token after refresh")
	}
	return creds.AccessToken, nil
}

// Client is a minimal Microsoft Graph HTTP helper.
type Client struct {
	Dir    string
	HTTP   *http.Client
	TZName string // IANA timezone for Prefer: outlook.timezone
	// AccessToken, if non-empty, is used as the Bearer token for every request.
	// No refresh and no credentials.json are used (multi-user / ephemeral token mode).
	AccessToken string
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (c *Client) authHeader(ctx context.Context) (string, error) {
	if tok := strings.TrimSpace(c.AccessToken); tok != "" {
		return tok, nil
	}
	if strings.TrimSpace(c.Dir) == "" {
		return "", fmt.Errorf("no access token: set --access-token or OUTLOOK_ACCESS_TOKEN, or use config under ~/.outlook-mcp")
	}
	return EnsureAccessToken(ctx, c.Dir)
}

func (c *Client) graphGET(ctx context.Context, u string) ([]byte, int, error) {
	token, err := c.authHeader(ctx)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if c.tzPrefer() != "" && (strings.Contains(u, "/calendar") || strings.Contains(u, "calendarView")) {
		req.Header.Set("Prefer", `outlook.timezone="`+c.tzPrefer()+`"`)
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}

func (c *Client) tzPrefer() string {
	if c.TZName != "" {
		return c.TZName
	}
	return "UTC"
}

func (c *Client) graphJSON(ctx context.Context, method, u string, body io.Reader, contentType string) ([]byte, int, error) {
	token, err := c.authHeader(ctx)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if strings.Contains(u, "graph.microsoft.com") && c.tzPrefer() != "" &&
		(strings.Contains(u, "/calendar") || strings.Contains(u, "calendarView")) {
		req.Header.Set("Prefer", `outlook.timezone="`+c.tzPrefer()+`"`)
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}
