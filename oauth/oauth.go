// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/xmidt-org/vouch/events"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	OAUTH2_TYPE = "oauth2"
)

type Config struct {
	// Priority is the priority of this config relative to others.
	// Higher numbers are higher priority. 0 means default, which is 1000.
	Priority int

	// ClientID is the application's ID.
	ClientID string

	// ClientSecret is the application's secret.
	ClientSecret string

	// TokenURL is the resource server's token endpoint URL.
	TokenURL string

	// Scopes specifies optional requested permissions.
	Scopes []string

	// EndpointParams specifies additional parameters for requests to the token endpoint.
	EndpointParams url.Values

	// AuthStyle optionally specifies how the endpoint wants the
	// client ID & client secret sent.
	// Valid values: ""/"auto_detect", "in_params", "in_header"
	AuthStyle string

	// ExpirationSafetyMargin is the percentage of the token lifetime
	// to use as a safety margin for token refresh. (0.8 = 20% early refresh)
	ExpirationSafetyMargin float64

	// DefaultTokenDuration is the assumed token lifetime if no expiry
	// is provided by the token endpoint. 0 or less = no expiry.
	DefaultTokenDuration time.Duration
}

func (c *Config) IsActive() bool {
	return c.ClientID != "" && c.TokenURL != ""
}

type OAuth struct {
	ts       oauth2.TokenSource
	dispatch func(any)
	priority int
}

func (o *OAuth) Priority() int {
	return o.priority
}

var styleMap = map[string]oauth2.AuthStyle{
	"":            oauth2.AuthStyleAutoDetect,
	"auto_detect": oauth2.AuthStyleAutoDetect,
	"in_params":   oauth2.AuthStyleInParams,
	"in_header":   oauth2.AuthStyleInHeader,
}

// New creates an OAuth client that automatically refreshes tokens safely.
func New(config Config, dispatch func(any)) (*OAuth, error) {
	style, ok := styleMap[config.AuthStyle]
	if !ok {
		return nil, fmt.Errorf("invalid AuthStyle: %s", config.AuthStyle)
	}

	if config.ExpirationSafetyMargin < 0 || config.ExpirationSafetyMargin > 1 {
		return nil, fmt.Errorf("ExpirationSafetyMargin must be between 0 and 1")
	}

	priority := config.Priority
	if config.Priority == 0 {
		priority = 1000
	}

	cfg := clientcredentials.Config{
		ClientID:       config.ClientID,
		ClientSecret:   config.ClientSecret,
		TokenURL:       config.TokenURL,
		Scopes:         config.Scopes,
		EndpointParams: config.EndpointParams,
		AuthStyle:      style,
	}

	baseTS := cfg.TokenSource(context.Background())

	if dispatch == nil {
		dispatch = func(any) {}
	}

	// Wrap base source to inject synthetic expiry if needed
	safeTS := &safetyMarginTokenSource{
		source:               baseTS,
		safetyMargin:         config.ExpirationSafetyMargin,
		defaultTokenLifetime: config.DefaultTokenDuration,
		dispatch:             dispatch,
	}

	// Wrap again with ReuseTokenSource for caching and refresh
	reuseTS := oauth2.ReuseTokenSource(nil, safeTS)

	return &OAuth{
		ts:       reuseTS,
		dispatch: dispatch,
		priority: priority,
	}, nil
}

// Decorate sets the Authorization header on an outgoing request.
func (o *OAuth) Decorate(req *http.Request) error {
	evnt := events.DecorateEvent{
		At:   time.Now(),
		Type: OAUTH2_TYPE,
	}
	token, err := o.ts.Token()
	evnt.Duration = time.Since(evnt.At)
	if err != nil {
		evnt.Err = err
		o.dispatch(evnt)
		return err
	}
	evnt.Expiration = token.Expiry

	token.SetAuthHeader(req)

	o.dispatch(evnt)
	return nil
}

// safetyMarginTokenSource ensures token has a safe expiry with a margin.
type safetyMarginTokenSource struct {
	source               oauth2.TokenSource
	safetyMargin         float64
	defaultTokenLifetime time.Duration
	dispatch             func(any)

	mu           sync.Mutex
	lastToken    *oauth2.Token
	lastAdjusted bool
}

func (s *safetyMarginTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	evnt := events.FetchEvent{
		At: time.Now(),
	}

	tok, err := s.source.Token()
	if err != nil {
		evnt.Err = err
		s.dispatch(evnt)
		return nil, err
	}
	evnt.Duration = time.Since(evnt.At)

	// If this is the same token we already adjusted, skip re-adjustment
	if tok == s.lastToken && s.lastAdjusted {
		evnt.Expiration = tok.Expiry
		s.dispatch(evnt)
		return tok, nil
	}

	issuedAt := time.Now()

	var lifetime time.Duration
	if tok.ExpiresIn > 0 {
		lifetime = time.Second * time.Duration(tok.ExpiresIn)
	} else {
		lifetime = s.defaultTokenLifetime
	}

	if lifetime > 0 {
		evnt.OriginalExpiration = issuedAt.Add(lifetime)
		safeLifetime := time.Duration(float64(lifetime) * s.safetyMargin)
		newExpiry := issuedAt.Add(safeLifetime)
		tok.Expiry = newExpiry
	}

	s.lastToken = tok
	s.lastAdjusted = true

	evnt.Expiration = tok.Expiry
	s.dispatch(evnt)
	return tok, nil
}
