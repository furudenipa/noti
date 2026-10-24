// SPDX-FileCopyrightText: 2026 Noti Contributors
//
// SPDX-License-Identifier: MPL-2.0

package discord

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbedValidation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		title   string
		message string
		want    string
	}{
		{"empty", "", "", "empty embed"},
		{"whitespace only", " ", "\n\t", "empty embed"},
		{"title too long", strings.Repeat("あ", 257), "message", "title exceeds 256"},
		{"message too long", "title", strings.Repeat("あ", 4097), "description exceeds 4096"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := EmbedNotification{WebhookURL: "http://example.invalid/secret", Title: tc.title, Message: tc.message}
			if err := n.Send(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %v, want %s", err, tc.want)
			}
		})
	}
}

func TestEmbedBoundaries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer ts.Close()
	for _, tc := range []struct{ name, title, message string }{
		{"maximum length", strings.Repeat("あ", 256), strings.Repeat("あ", 4096)},
		{"title only", "title", ""},
		{"message only", "", "message"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := EmbedNotification{WebhookURL: ts.URL, Title: tc.title, Message: tc.message}
			if err := n.Send(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
