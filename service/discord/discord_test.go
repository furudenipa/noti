// SPDX-FileCopyrightText: 2026 Noti Contributors
//
// SPDX-License-Identifier: MPL-2.0

package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		url       string
		message   string
		wantError string
	}{
		{"missing URL", "", "message", "missing webhook URL"},
		{"invalid URL", "://secret", "message", "invalid webhook URL"},
		{"relative URL", "/webhook/secret", "message", "invalid webhook URL"},
		{"unsupported scheme", "ftp://example.com/secret", "message", "invalid webhook URL"},
		{"missing content", "https://example.com/secret", "", "missing content"},
		{"too long", "https://example.com/secret", strings.Repeat("あ", 2001), "exceeds 2000 characters"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := Notification{WebhookURL: tc.url, Message: tc.message}
			err := n.Send()
			if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("got error %v, want %q", err, tc.wantError)
			}
		})
	}
}

func TestSendDefaultsAndUnicodeLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if len(payload) != 2 {
			t.Errorf("optional fields should be omitted: %v", payload)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	n := Notification{WebhookURL: ts.URL, Message: strings.Repeat("あ", 2000)}
	if err := n.Send(); err != nil {
		t.Fatal(err)
	}
}

func TestSendContent(t *testing.T) {
	for _, tc := range []struct{ name, title, message, want string }{
		{"title and message", "Build", "Done!", "**Build**\nDone!"},
		{"title only", "Build", "", "**Build**"},
		{"message only", "", "Done!", "Done!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload struct {
					Content string `json:"content"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if payload.Content != tc.want {
					t.Errorf("got content %q, want %q", payload.Content, tc.want)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer ts.Close()
			n := Notification{WebhookURL: ts.URL, Title: tc.title, Message: tc.message}
			if err := n.Send(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSendFormattedContentLimit(t *testing.T) {
	n := Notification{WebhookURL: "https://example.invalid/unused", Title: "A", Message: strings.Repeat("あ", 1995)}
	// The bold markers and newline count toward the 2000-character limit.
	if err := n.Send(); err == nil || !strings.Contains(err.Error(), "exceeds 2000 characters") {
		t.Errorf("unexpected error: %v", err)
	}
}
