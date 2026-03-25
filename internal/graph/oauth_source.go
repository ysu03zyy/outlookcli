package graph

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/ysu03zyy/outlookcli/internal/config"
)

const tokenExchangeTimeout = 30 * time.Second

var microsoftEndpoint = oauth2.Endpoint{
	TokenURL: tokenEndpoint,
}

func scopeSlice() []string {
	return strings.Fields(graphScope)
}

// tokenFromRaw builds an oauth2.Token from credentials.json content.
func tokenFromRaw(raw map[string]any) *oauth2.Token {
	t := &oauth2.Token{}
	if v, ok := raw["access_token"].(string); ok {
		t.AccessToken = v
	}
	if v, ok := raw["refresh_token"].(string); ok {
		t.RefreshToken = v
	}
	if v, ok := raw["expiry"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, v); err == nil {
			t.Expiry = parsed
		} else if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			t.Expiry = parsed
		}
	}
	return t
}

func mergeTokenIntoRaw(raw map[string]any, tok *oauth2.Token) {
	if tok == nil {
		return
	}
	raw["access_token"] = tok.AccessToken
	if tok.RefreshToken != "" {
		raw["refresh_token"] = tok.RefreshToken
	}
	if tok.TokenType != "" {
		raw["token_type"] = tok.TokenType
	}
	if !tok.Expiry.IsZero() {
		raw["expiry"] = tok.Expiry.UTC().Format(time.RFC3339Nano)
	}
}

type persistingTokenSource struct {
	base oauth2.TokenSource
	dir  string
	raw  map[string]any

	mu sync.Mutex
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	t, err := p.base.Token()
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	prevAT, _ := p.raw["access_token"].(string)
	prevRT, _ := p.raw["refresh_token"].(string)
	mergeTokenIntoRaw(p.raw, t)
	if t.AccessToken == prevAT && strings.TrimSpace(t.RefreshToken) == strings.TrimSpace(prevRT) {
		return t, nil
	}
	if err := SaveCredentials(p.dir, p.raw); err != nil {
		return nil, fmt.Errorf("persist token: %w", err)
	}
	return t, nil
}

// BuildTokenSource returns an oauth2.TokenSource that refreshes only when needed
// and persists updated tokens to credentials.json (including expiry for lazy refresh).
func BuildTokenSource(ctx context.Context, dir string) (oauth2.TokenSource, error) {
	app, err := config.LoadAppConfig(dir)
	if err != nil {
		return nil, err
	}
	_, raw, err := LoadCredentials(dir)
	if err != nil {
		return nil, err
	}
	init := tokenFromRaw(raw)
	if init.RefreshToken == "" {
		return nil, fmt.Errorf("credentials.json: missing refresh_token; run Outlook setup (see README)")
	}

	cfg := &oauth2.Config{
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		Endpoint:     microsoftEndpoint,
		Scopes:       scopeSlice(),
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: tokenExchangeTimeout})
	base := cfg.TokenSource(ctx, init)
	return &persistingTokenSource{base: base, dir: dir, raw: raw}, nil
}
