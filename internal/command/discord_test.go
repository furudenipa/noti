// SPDX-FileCopyrightText: 2026 Noti Contributors
//
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestDiscordSelection(t *testing.T) {
	for _, tc := range []struct {
		name     string
		defaults []string
		env      string
		flags    []string
		want     string
	}{
		{"file default", []string{"discord"}, "", nil, "discord"},
		{"environment default", []string{"banner"}, "discord", nil, "discord"},
		{"flag overrides defaults", []string{"banner"}, "slack", []string{"--discord"}, "discord"},
		{"disable flag overrides defaults", []string{"discord"}, "discord", []string{"--discord=false"}, ""},
		{"other flag overrides discord", []string{"discord"}, "discord", []string{"--slack"}, "slack"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NOTI_DEFAULT", tc.env)
			v := viper.New()
			v.Set("defaults", tc.defaults)
			flags := pflag.NewFlagSet("discord", pflag.ContinueOnError)
			InitFlags(flags)
			if err := flags.Parse(tc.flags); err != nil {
				t.Fatal(err)
			}
			services := enabledServices(v, flags)
			if tc.want == "" {
				if len(services) != 0 {
					t.Errorf("unexpected services: %v", services)
				}
			} else if _, ok := services[tc.want]; !ok || len(services) != 1 {
				t.Errorf("got services %v, want %s only", services, tc.want)
			}
		})
	}
}

func TestDiscordConfiguredDelivery(t *testing.T) {
	for _, source := range []string{"yaml", "environment"} {
		t.Run(source, func(t *testing.T) {
			var hit bool
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hit = true
				var payload struct {
					Content   string `json:"content"`
					Username  string `json:"username"`
					AvatarURL string `json:"avatar_url"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				if payload.Content != "**Build**\nDone!" || payload.Username != "builder" ||
					payload.AvatarURL != "https://example.com/icon.png" {
					t.Errorf("unexpected payload: %+v", payload)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer ts.Close()
			t.Setenv("NOTI_DEFAULT", "")
			for _, env := range []string{"NOTI_DISCORD_WEBHOOKURL", "NOTI_DISCORD_USERNAME", "NOTI_DISCORD_AVATARURL"} {
				t.Setenv(env, "")
			}
			v := viper.New()
			setNotiDefaults(v)
			if err := bindNotiEnv(v); err != nil {
				t.Fatal(err)
			}
			v.SetConfigType("yaml")
			config := "defaults: [discord]\ndiscord:\n  webhookURL: " + ts.URL +
				"\n  username: builder\n  avatarURL: https://example.com/icon.png\n"
			if err := v.ReadConfig(strings.NewReader(config)); err != nil {
				t.Fatal(err)
			}
			if source == "environment" {
				// Ensure environment settings override YAML, not just defaults.
				config = "defaults: [discord]\ndiscord:\n  webhookURL: https://example.invalid/unused" +
					"\n  username: unused\n  avatarURL: https://example.invalid/unused\n"
				if err := v.ReadConfig(strings.NewReader(config)); err != nil {
					t.Fatal(err)
				}
				t.Setenv("NOTI_DISCORD_WEBHOOKURL", ts.URL)
				t.Setenv("NOTI_DISCORD_USERNAME", "builder")
				t.Setenv("NOTI_DISCORD_AVATARURL", "https://example.com/icon.png")
			}
			v.Set("title", "Build")
			v.Set("message", "Done!")
			flags := pflag.NewFlagSet("discord", pflag.ContinueOnError)
			InitFlags(flags)
			notis := getNotifications(v, enabledServices(v, flags))
			if len(notis) != 1 {
				t.Fatalf("got %d notifications, want 1", len(notis))
			}
			if err := notis[0].Send(); err != nil {
				t.Fatal(err)
			}
			if !hit {
				t.Error("did not reach webhook")
			}
		})
	}
}
