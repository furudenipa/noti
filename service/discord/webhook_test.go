// SPDX-FileCopyrightText: 2026 Noti Contributors
//
// SPDX-License-Identifier: MPL-2.0

package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestSend(t *testing.T) {
	for _, format := range []struct {
		name    string
		send    func(string, *http.Client) error
		payload string
	}{
		{"plain", func(url string, client *http.Client) error {
			n := Notification{WebhookURL: url, Client: client, Title: "完了", Message: "@everyone <@123> \"test\"", Username: "noti", AvatarURL: "https://example.com/icon.png"}
			return n.Send()
		}, `{"content":"**完了**\n@everyone <@123> \"test\"","username":"noti","avatar_url":"https://example.com/icon.png","allowed_mentions":{"parse":[]}}`},
		{"embed", func(url string, client *http.Client) error {
			n := EmbedNotification{WebhookURL: url, Client: client, Title: "完了", Message: "@everyone <@123> \"test\"", Username: "noti", AvatarURL: "https://example.com/icon.png"}
			return n.Send()
		}, `{"embeds":[{"title":"完了","description":"@everyone <@123> \"test\""}],"username":"noti","avatar_url":"https://example.com/icon.png","allowed_mentions":{"parse":[]}}`},
	} {
		t.Run(format.name, func(t *testing.T) {
			for _, tc := range []struct {
				name            string
				status          int
				body, wantError string
			}{
				{"message", 200, `{"id":"123"}`, ""},
				{"no content", 204, "", ""},
				{"bad request", 400, `{"code":50035,"message":"Invalid Form Body"}`, "HTTP 400 (code 50035): Invalid Form Body"},
				{"unknown webhook", 404, `{"code":10015,"message":"Unknown Webhook"}`, "HTTP 404 (code 10015): Unknown Webhook"},
				{"rate limit", 429, `{"message":"You are being rate limited.","retry_after":1}`, "HTTP 429"},
				{"server error", 500, "<html>error</html>", "HTTP status 500"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var want map[string]interface{}
					if err := json.Unmarshal([]byte(format.payload), &want); err != nil {
						t.Fatal(err)
					}
					ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
							t.Errorf("unexpected request: %s, %s", r.Method, r.Header.Get("Content-Type"))
						}
						if r.URL.Query().Get("wait") != "true" || r.URL.Query().Get("thread_id") != "123" {
							t.Errorf("unexpected query: %s", r.URL.RawQuery)
						}
						var got map[string]interface{}
						if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
							t.Error(err)
							w.WriteHeader(http.StatusBadRequest)
							return
						}
						if !reflect.DeepEqual(got, want) {
							t.Errorf("got payload %v, want %v", got, want)
						}
						w.WriteHeader(tc.status)
						w.Write([]byte(tc.body))
					}))
					defer ts.Close()
					err := format.send(ts.URL+"/api/webhooks/id/token?thread_id=123&wait=false", ts.Client())
					if tc.wantError == "" {
						if err != nil {
							t.Fatal(err)
						}
					} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
						t.Errorf("got error %v, want %q", err, tc.wantError)
					}
				})
			}
		})
	}
}

func TestSendTransportError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()
	for name, n := range map[string]interface{ Send() error }{
		"plain": &Notification{WebhookURL: ts.URL + "/secret-token", Client: ts.Client(), Message: "message"},
		"embed": &EmbedNotification{WebhookURL: ts.URL + "/secret-token", Client: ts.Client(), Title: "title"},
	} {
		t.Run(name, func(t *testing.T) {
			err := n.Send()
			if err == nil {
				t.Fatal("expected connection error")
			}
			if strings.Contains(err.Error(), "secret-token") {
				t.Errorf("webhook token leaked: %v", err)
			}
		})
	}
}
