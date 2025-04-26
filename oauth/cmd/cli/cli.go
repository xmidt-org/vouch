// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xmidt-org/vouch/events"
	"github.com/xmidt-org/vouch/oauth"
)

func main() {
	var scopes []string
	if os.Getenv("OAUTH_SCOPES") != "" {
		for _, item := range strings.Split(os.Getenv("OAUTH_SCOPES"), ",") {
			scopes = append(scopes, strings.TrimSpace(item))
		}
	}
	config := oauth.Config{
		TokenURL:     strings.TrimSpace(os.Getenv("OAUTH_URL")),
		ClientID:     strings.TrimSpace(os.Getenv("OAUTH_CLIENT_ID")),
		ClientSecret: strings.TrimSpace(os.Getenv("OAUTH_CLIENT_SECRET")),
		Scopes:       scopes,
	}

	o, err := oauth.New(config, listen)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		panic(err)
	}

	err = o.Decorate(req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Request Headers: %v\n", req.Header)
}

func listen(evnt any) {
	switch e := evnt.(type) {
	case events.FetchEvent:
		fmt.Println("Fetch Event:")
		fmt.Printf("  At:         %s\n", e.At.Format(time.RFC3339))
		fmt.Printf("  Type:       %s\n", e.Type)
		fmt.Printf("  Duration:   %s\n", e.Duration)
		fmt.Printf("  Expiration: %s\n", e.Expiration.Format(time.RFC3339))
		fmt.Printf("  Error:      %v\n", e.Err)
	case events.DecorateEvent:
		fmt.Println("Decorate Event:")
		fmt.Printf("  At:         %s\n", e.At.Format(time.RFC3339))
		fmt.Printf("  Type:       %s\n", e.Type)
		fmt.Printf("  Duration:   %s\n", e.Duration)
		fmt.Printf("  Expiration: %s\n", e.Expiration.Format(time.RFC3339))
		fmt.Printf("  Error:      %v\n", e.Err)
	default:
		fmt.Println("Unknown Event:", e)
	}
}
