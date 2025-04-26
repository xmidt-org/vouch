// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package vouch

import (
	"sort"

	"github.com/xmidt-org/vouch/basic"
	"github.com/xmidt-org/vouch/events"
	"github.com/xmidt-org/vouch/oauth"
)

type optFuncErr func(*Vouch) error

func (f optFuncErr) apply(a *Vouch) error {
	return f(a)
}

type optFunc func(*Vouch)

func (f optFunc) apply(a *Vouch) error {
	f(a)
	return nil
}

// WithFetchEventListener adds a FetchEventListener to the Auth instance.
// It returns a cancel function that can be used to remove the listener.
// If the listener is nil, it does nothing.
func WithFetchEventListener(listener events.FetchEventListener, cancel ...*func()) Option {
	return optFunc(func(a *Vouch) {
		if listener == nil {
			return
		}
		c := a.fetchListeners.Add(listener)
		for _, cancelFunc := range cancel {
			if cancelFunc != nil {
				*cancelFunc = c
			}
		}
	})
}

// WithDecorateEventListener adds a DecorateEventListener to the Auth instance.
// It returns a cancel function that can be used to remove the listener.
// If the listener is nil, it does nothing.
func WithDecorateEventListener(listener events.DecorateEventListener, cancel ...*func()) Option {
	return optFunc(func(a *Vouch) {
		if listener == nil {
			return
		}
		c := a.decorateListeners.Add(listener)
		for _, cancelFunc := range cancel {
			if cancelFunc != nil {
				*cancelFunc = c
			}
		}
	})
}

// -----------------------------------------------------------------------------

func handleBasic(b basic.Config) Option {
	return optFunc(func(a *Vouch) {
		a.basic = basic.New(b, a.dispatch)
	})
}

func handleOAuth(o oauth.Config) Option {
	return optFuncErr(func(a *Vouch) error {
		var err error
		a.oauth, err = oauth.New(o, a.dispatch)
		return err
	})
}

func setupDecorators() Option {
	return optFunc(func(a *Vouch) {
		if a.basic != nil {
			a.decorators = append(a.decorators, a.basic)
		}
		if a.oauth != nil {
			a.decorators = append(a.decorators, a.oauth)
		}

		sort.Slice(a.decorators, func(i, j int) bool {
			return a.decorators[i].Priority() > a.decorators[j].Priority()
		})
	})
}
