// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package vouch

import (
	"errors"
	"net/http"
	"testing"

	"github.com/xmidt-org/vouch/basic"
	"github.com/xmidt-org/vouch/events"
	"github.com/xmidt-org/vouch/oauth"
)

func TestVouch_New(t *testing.T) {
	tests := []struct {
		description string
		config      Config
		options     []Option
		expectError bool
	}{
		{
			description: "Valid configuration with both Basic and OAuth",
			config: Config{
				Basic: basic.Config{
					Username: "test-user",
					Password: "test-pass",
				},
				OAuth: oauth.Config{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					TokenURL:     "https://example.com/token",
				},
			},
			expectError: false,
		},
		{
			description: "Empty configuration",
			config:      Config{},
			expectError: false,
		},
		{
			description: "Invalid OAuth configuration",
			config: Config{
				OAuth: oauth.Config{
					AuthStyle: "invalid",
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			_, err := New(tc.config, tc.options...)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got: %v", err)
				}
			}
		})
	}
}

func TestVouch_Decorate(t *testing.T) {
	tests := []struct {
		description string
		decorators  []decorator
		expectError bool
	}{
		{
			description: "No decorators",
			decorators:  nil,
			expectError: false,
		},
		{
			description: "Single successful decorator",
			decorators: []decorator{
				mockDecorator{
					decorateFunc: func(req *http.Request) error {
						return nil
					},
				},
			},
			expectError: false,
		},
		{
			description: "All decorators fail",
			decorators: []decorator{
				mockDecorator{
					decorateFunc: func(req *http.Request) error {
						return errors.New("failed")
					},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			v := &Vouch{
				decorators: tc.decorators,
			}

			req, _ := http.NewRequest("GET", "https://example.com", nil)
			err := v.Decorate(req)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got: %v", err)
				}
			}
		})
	}
}

func TestVouch_AddFetchEventListener(t *testing.T) {
	listenerCalled := false
	listener := events.FetchEventListenerFunc(func(event events.FetchEvent) {
		listenerCalled = true
	})

	v := &Vouch{}
	cancel := v.AddFetchEventListener(listener)

	// Trigger a fetch event
	event := events.FetchEvent{}
	v.dispatch(event)

	if !listenerCalled {
		t.Errorf("expected listener to be called, but it was not")
	}

	// Test canceling the listener
	listenerCalled = false
	cancel()
	v.dispatch(event)

	if listenerCalled {
		t.Errorf("expected listener to be canceled, but it was still called")
	}
}

func TestVouch_AddDecorateEventListener(t *testing.T) {
	listenerCalled := false
	listener := events.DecorateEventListenerFunc(func(event events.DecorateEvent) {
		listenerCalled = true
	})

	v := &Vouch{}
	cancel := v.AddDecorateEventListener(listener)

	// Trigger a decorate event
	event := events.DecorateEvent{}
	v.dispatch(event)

	if !listenerCalled {
		t.Errorf("expected listener to be called, but it was not")
	}

	// Test canceling the listener
	listenerCalled = false
	cancel()
	v.dispatch(event)

	if listenerCalled {
		t.Errorf("expected listener to be canceled, but it was still called")
	}
}

type mockDecorator struct {
	decorateFunc func(*http.Request) error
}

func (m mockDecorator) Decorate(req *http.Request) error {
	return m.decorateFunc(req)
}

func (m mockDecorator) Priority() int {
	return 0
}
