// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"time"
)

// FetchEvent is the event that is sent about a fetch request.
type FetchEvent struct {
	// At holds the time when the fetch request was made.
	At time.Time

	// Type is the type of token that was fetched.
	// It can be "oauth2" or "basic".
	Type string

	// Duration is the time waited for the token/response.
	Duration time.Duration

	// Expiration is the time the token expires.
	Expiration time.Time

	// OriginalExpiration is the original expiration time of the token, before
	// any adjustments made by the client.
	OriginalExpiration time.Time

	// Error is the error returned from the OAuth service.
	Err error
}

// FetchEventListener is the interface that must be implemented by types that
// want to receive FetchEvent notifications.
type FetchEventListener interface {
	OnFetchEvent(FetchEvent)
}

// FetchEventListenerFunc is a function type that implements FetchEventListener.
// It can be used as an adapter for functions that need to implement the
// FetchEventListener interface.
type FetchEventListenerFunc func(FetchEvent)

func (f FetchEventListenerFunc) OnFetchEvent(e FetchEvent) {
	f(e)
}

// DecorateEvent is the event that is sent about a decorate request.
type DecorateEvent struct {
	// At holds the time when the fetch request was made.
	At time.Time

	// Type is the type of token that was fetched.
	// It can be "oauth2" or "basic".
	Type string

	// Duration is the time waited for the token/response.
	Duration time.Duration

	// Expiration is the time the token expires.
	Expiration time.Time

	// Error is the error returned from the OAuth service.
	Err error
}

// DecorateEventListener is the interface that must be implemented by types that
// want to receive DecorateEvent notifications.
type DecorateEventListener interface {
	OnDecorateEvent(DecorateEvent)
}

// DecorateEventListenerFunc is a function type that implements DecorateEventListener.
// It can be used as an adapter for functions that need to implement the
// DecorateEventListener interface.
type DecorateEventListenerFunc func(DecorateEvent)

func (f DecorateEventListenerFunc) OnDecorateEvent(e DecorateEvent) {
	f(e)
}
