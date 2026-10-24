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

func TestSend(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		wantError string
	}{
		{"message", 200, `{"id":"123"}`, ""},
		{"no content", 204, "", ""},
		{"bad request", 400, `{"code":50035,"message":"Invalid Form Body"}`, "HTTP 400 (code 50035): Invalid Form Body"},
		{"unknown webhook", 404, `{"code":10015,"message":"Unknown Webhook"}`, "HTTP 404 (code 10015): Unknown Webhook"},
		{"rate limit", 429, `{"message":"You are being rate limited.","retry_after":1}`, "HTTP 429"},
		{"server error", 500, "<html>error</html>", "HTTP status 500"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var hit bool
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hit = true
				if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("unexpected request: %s, %s", r.Method, r.Header.Get("Content-Type"))
				}
				if r.URL.Query().Get("wait") != "true" || r.URL.Query().Get("thread_id") != "123" {
					t.Errorf("unexpected query: %s", r.URL.RawQuery)
				}
				var payload map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				for key, want := range map[string]string{
					"content":    "**完了**\n@everyone <@123> \"test\"",
					"username":   "noti",
					"avatar_url": "https://example.com/icon.png",
				} {
					var got string
					if err := json.Unmarshal(payload[key], &got); err != nil || got != want {
						t.Errorf("%s: got %q, want %q (error: %v)", key, got, want, err)
					}
				}
				if string(payload["allowed_mentions"]) != `{"parse":[]}` {
					t.Errorf("mentions not disabled: %s", payload["allowed_mentions"])
				}
				if len(payload) != 4 {
					t.Errorf("unexpected payload fields: %v", payload)
				}
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer ts.Close()
			n := Notification{
				WebhookURL: ts.URL + "/api/webhooks/id/token?thread_id=123&wait=false",
				Content:    "**完了**\n@everyone <@123> \"test\"",
				Username:   "noti",
				AvatarURL:  "https://example.com/icon.png",
				Client:     ts.Client(),
			}
			err := n.Send()
			if tc.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("got error %v, want %q", err, tc.wantError)
			}
			if !hit {
				t.Error("did not reach webhook")
			}
		})
	}
}

func TestSendValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		url       string
		content   string
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
			n := Notification{WebhookURL: tc.url, Content: tc.content}
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
	n := Notification{WebhookURL: ts.URL, Content: strings.Repeat("あ", 2000)}
	if err := n.Send(); err != nil {
		t.Fatal(err)
	}
}

func TestSendTransportError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()
	n := Notification{WebhookURL: ts.URL + "/secret-token", Content: "message", Client: ts.Client()}
	err := n.Send()
	if err == nil {
		t.Fatal("expected connection error")
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Errorf("webhook token leaked: %v", err)
	}
}
