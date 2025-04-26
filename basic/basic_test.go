// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/vouch/events"
)

func TestNew(t *testing.T) {
	tests := []struct {
		description string
		config      Config
		want        Basic
		withFn      bool
		header      string
	}{
		{
			description: "empty configuration means no auth",
		}, {
			description: "a simple configuration",
			config: Config{
				Username: "user",
				Password: "pass",
			},
			want: Basic{
				priority: 300,
				username: "user",
				password: "pass",
			},
			header: "Basic dXNlcjpwYXNz",
		}, {
			description: "a configuration with events and priority",
			config: Config{
				Priority: 12,
				Username: "user",
				Password: "pass",
			},
			want: Basic{
				priority: 12,
				username: "user",
				password: "pass",
			},
			header: "Basic dXNlcjpwYXNz",
			withFn: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			want := tc.want
			cfg := tc.config

			fn := func(evnt any) {
				switch e := evnt.(type) {
				case events.DecorateEvent:
					assert.NotZero(e.At)
					assert.Equal(BASIC_TYPE, e.Type)
					assert.NoError(e.Err)
					return
				}
				assert.Fail("unexpected event type")
			}

			if !tc.withFn {
				fn = nil
			}

			got := New(cfg, fn)

			if tc.config.Username == "" && tc.config.Password == "" {
				assert.Nil(got)
				return
			}

			assert.Equal(want.priority, got.priority)
			assert.Equal(want.username, got.username)
			assert.Equal(want.password, got.password)

			req, err := http.NewRequest("GET", "http://example.com", nil)
			require.NoError(err)

			err = got.Decorate(req)
			assert.NoError(err)

			assert.Equal(tc.header, req.Header.Get("Authorization"))

			assert.Equal(want.priority, got.Priority())
		})
	}
}
