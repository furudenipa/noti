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

func TestEmbedColor(t *testing.T) {
	for _, tc := range []struct {
		name   string
		failed bool
		want   int
	}{
		{"success", false, 0x2ECC71},
		{"failure", true, 0xE74C3C},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload struct {
					Embeds []struct {
						Color int `json:"color"`
					} `json:"embeds"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if len(payload.Embeds) != 1 {
					t.Error("expected one embed")
				} else if payload.Embeds[0].Color != tc.want {
					t.Errorf("got color %#x, want %#x", payload.Embeds[0].Color, tc.want)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer ts.Close()
			n := EmbedNotification{WebhookURL: ts.URL, Title: "Build", Message: "same message", Failed: tc.failed}
			if err := n.Send(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
