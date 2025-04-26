// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"net/http"
	"time"

	"github.com/xmidt-org/vouch/events"
)

const (
	BASIC_TYPE = "basic"
)

type Config struct {
	// Priority is the priority of this config relative to others.
	// Higher numbers are higher priority. 0 means default, which is 300.
	Priority int

	// Username is the username for basic authentication.
	Username string

	// Password is the password for basic authentication.
	Password string
}

func (c *Config) IsActive() bool {
	return c.Username != ""
}

type Basic struct {
	priority int
	username string
	password string
	dispatch func(any)
}

func (b *Basic) Priority() int {
	return b.priority
}

func New(config Config, dispatch func(any)) *Basic {
	if !config.IsActive() {
		return nil
	}

	priority := config.Priority
	if priority == 0 {
		priority = 300
	}
	if dispatch == nil {
		dispatch = func(any) {}
	}
	return &Basic{
		username: config.Username,
		password: config.Password,
		priority: priority,
		dispatch: dispatch,
	}
}

func (b *Basic) Decorate(req *http.Request) error {
	evnt := events.DecorateEvent{
		At:   time.Now(),
		Type: BASIC_TYPE,
	}
	req.SetBasicAuth(b.username, b.password)
	b.dispatch(evnt)
	return nil
}
