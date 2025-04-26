// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		description string
		config      Config
		dispatch    func(any)
		expectError bool
	}{
		{
			description: "Valid configuration",
			config: Config{
				ClientID:               "test-client",
				ClientSecret:           "test-secret",
				TokenURL:               "https://example.com/token",
				ExpirationSafetyMargin: 0.8,
			},
			dispatch:    func(any) {},
			expectError: false,
		},
		{
			description: "Invalid AuthStyle",
			config: Config{
				ClientID:  "test-client",
				TokenURL:  "https://example.com/token",
				AuthStyle: "invalid",
			},
			dispatch:    func(any) {},
			expectError: true,
		},
		{
			description: "Invalid ExpirationSafetyMargin",
			config: Config{
				ClientID:               "test-client",
				TokenURL:               "https://example.com/token",
				ExpirationSafetyMargin: 1.5,
			},
			dispatch:    func(any) {},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			_, err := New(tc.config, tc.dispatch)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOAuthEndToEnd(t *testing.T) {
	tests := []struct {
		description string
		config      Config
		mockServer  func() string // Returns the mock server URL
		expectError bool
	}{
		{
			description: "Successful token retrieval",
			config: Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				TokenURL:     "mock-url-will-be-replaced",
			},
			mockServer: func() string {
				// Start a mock server to simulate token endpoint
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"access_token": "mock-token", "expires_in": 3600}`))
				}))
				return server.URL
			},
			expectError: false,
		},
		{
			description: "Invalid token response",
			config: Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				TokenURL:     "mock-url-will-be-replaced",
			},
			mockServer: func() string {
				// Start a mock server to simulate token endpoint
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "invalid_request"}`))
				}))
				return server.URL
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			// Start the mock server
			mockURL := tc.mockServer()
			defer func() {
				if mockURL != "" {
					http.DefaultClient.CloseIdleConnections()
				}
			}()

			// Update the TokenURL to point to the mock server
			tc.config.TokenURL = mockURL

			// Perform the end-to-end test
			got, err := New(tc.config, func(any) {})
			require.NoError(t, err)

			req, err := http.NewRequest("GET", "https://example.com/resource", nil)
			require.NoError(t, err)

			err = got.Decorate(req)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
