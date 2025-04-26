// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package vouch

import (
	"errors"
	"net/http"
	"time"

	"github.com/xmidt-org/eventor"
	"github.com/xmidt-org/vouch/basic"
	"github.com/xmidt-org/vouch/events"
	"github.com/xmidt-org/vouch/oauth"
)

// Config is the configuration for the Vouch authentication system. It contains
// the configuration for both basic and OAuth authentication methods.
//
// Empty sub-structures are valid and disable the corresponding
// authentication method. For example, if Basic is empty, basic authentication
// is disabled. If OAuth is empty, OAuth authentication is disabled.
// If both Basic and OAuth are empty, no authentication decoration is performed.
type Config struct {
	// Basic is the configuration for basic authentication.
	Basic basic.Config

	// OAuth is the configuration for OAuth authentication.
	OAuth oauth.Config
}

type decorator interface {
	Decorate(*http.Request) error
	Priority() int
}

// Vouch is the main struct that holds the authentication methods and event
// listeners. It contains the OAuth and Basic authentication methods, as well as
// the event listeners for fetch and decorate events.
type Vouch struct {
	oauth *oauth.OAuth
	basic *basic.Basic

	fetchListeners    eventor.Eventor[events.FetchEventListener]
	decorateListeners eventor.Eventor[events.DecorateEventListener]

	decorators []decorator
}

// Option is a function that configures the Auth instance.
type Option interface {
	apply(*Vouch) error
}

// New creates a new Vouch instance with the given configuration and options.
func New(cfg Config, opts ...Option) (*Vouch, error) {
	var auth Vouch

	defaults := []Option{
		handleBasic(cfg.Basic),
		handleOAuth(cfg.OAuth),
	}

	finalize := []Option{
		setupDecorators(),
	}

	all := append(defaults, opts...)
	all = append(all, finalize...)

	for _, opt := range all {
		if err := opt.apply(&auth); err != nil {
			return nil, err
		}
	}

	return &auth, nil
}

// AddFetchEventListener adds a listener for fetch events.
func (a *Vouch) AddFetchEventListener(listener events.FetchEventListener) (cancel func()) {
	return a.fetchListeners.Add(listener)
}

// AddDecorateEventListener adds a listener for decorate events.
func (a *Vouch) AddDecorateEventListener(listener events.DecorateEventListener) (cancel func()) {
	return a.decorateListeners.Add(listener)
}

// Decorate decorates the request with the appropriate authentication method.
// It tries each decorator in order of priority until one succeeds or all fail.
func (a *Vouch) Decorate(req *http.Request) error {
	evnt := events.DecorateEvent{
		At:   time.Now(),
		Type: "none",
	}

	errs := make([]error, 0, len(a.decorators))
	for _, decorator := range a.decorators {
		err := decorator.Decorate(req)
		if err == nil {
			return nil
		}
		errs = append(errs, err)
	}

	err := errors.Join(errs...)
	evnt.Duration = time.Since(evnt.At)
	evnt.Err = err
	a.dispatch(evnt)
	return err
}

// dispatch dispatches the event to the listeners and returns the error that
// should be returned by the caller.
func (a *Vouch) dispatch(event any) {
	switch event := event.(type) {
	case events.FetchEvent:
		a.fetchListeners.Visit(func(listener events.FetchEventListener) {
			listener.OnFetchEvent(event)
		})
	case events.DecorateEvent:
		a.decorateListeners.Visit(func(listener events.DecorateEventListener) {
			listener.OnDecorateEvent(event)
		})
	}
}
