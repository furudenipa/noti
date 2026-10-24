// SPDX-FileCopyrightText: 2026 Noti Contributors
//
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
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
		{"both flags", nil, "", []string{"--discord", "--discord-embed"}, "discord discord-embed"},
		{"embed default", []string{"discord-embed"}, "", nil, "discord-embed"},
		{"embed environment", nil, "discord-embed", nil, "discord-embed"},
		{"embed flag", []string{"banner"}, "slack", []string{"--discord-embed"}, "discord-embed"},
		{"embed disabled", []string{"discord-embed"}, "", []string{"--discord-embed=false"}, ""},
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
			} else {
				want := strings.Fields(tc.want)
				if len(services) != len(want) {
					t.Errorf("got services %v, want %v", services, want)
				}
				for _, name := range want {
					if _, ok := services[name]; !ok {
						t.Errorf("missing service %s", name)
					}
				}
			}
		})
	}
}

func TestDiscordConfiguredDelivery(t *testing.T) {
	for _, tc := range []struct {
		services                 []string
		contentCount, embedCount int
	}{
		{[]string{"discord"}, 1, 0},
		{[]string{"discord-embed"}, 0, 1},
		{[]string{"discord", "discord-embed"}, 1, 1},
	} {
		mode := strings.Join(tc.services, ", ")
		for _, source := range []string{"yaml", "environment"} {
			t.Run(mode+"/"+source, func(t *testing.T) {
				var contentHits, embedHits int
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var payload struct {
						Content string `json:"content"`
						Embeds  []struct {
							Title       string
							Description string
						} `json:"embeds"`
						Username  string `json:"username"`
						AvatarURL string `json:"avatar_url"`
					}
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					if len(payload.Embeds) > 0 {
						embedHits++
						if payload.Content != "" || len(payload.Embeds) != 1 || payload.Embeds[0].Title != "Build" || payload.Embeds[0].Description != "Done!" {
							t.Errorf("unexpected embed: %+v", payload)
						}
					} else {
						contentHits++
						if payload.Content != "**Build**\nDone!" {
							t.Errorf("unexpected content: %+v", payload)
						}
					}
					if payload.Username != "builder" ||
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
				config := "defaults: [" + mode + "]\ndiscord:\n  webhookURL: " + ts.URL +
					"\n  username: builder\n  avatarURL: https://example.com/icon.png\n"
				if err := v.ReadConfig(strings.NewReader(config)); err != nil {
					t.Fatal(err)
				}
				if source == "environment" {
					// Ensure environment settings override YAML, not just defaults.
					config = "defaults: [" + mode + "]\ndiscord:\n  webhookURL: https://example.invalid/unused" +
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
				want := len(tc.services)
				if len(notis) != want {
					t.Fatalf("got %d notifications, want %d", len(notis), want)
				}
				for _, n := range notis {
					if err := n.Send(); err != nil {
						t.Fatal(err)
					}
				}
				if embedHits != tc.embedCount || contentHits != tc.contentCount {
					t.Errorf("unexpected formats: content=%d embed=%d", contentHits, embedHits)
				}
			})
		}
	}
}

func TestDiscordEmbedCommand(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is required")
	}
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			t.Setenv("NOTI_DEFAULT", "")
			for _, key := range []string{"NOTI_DISCORD_WEBHOOKURL", "NOTI_DISCORD_USERNAME", "NOTI_DISCORD_AVATARURL"} {
				t.Setenv(key, "")
			}
			var title, description string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload struct {
					Embeds []struct{ Title, Description string }
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				if len(payload.Embeds) == 1 {
					title = payload.Embeds[0].Title
					description = payload.Embeds[0].Description
				} else {
					t.Error("expected one embed")
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer ts.Close()
			config := filepath.Join(t.TempDir(), "noti.yaml")
			if err := os.WriteFile(config, []byte("defaults: [banner]\nmessage: Done!\ntime: true\ndiscord:\n  webhookURL: "+ts.URL+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := &cobra.Command{RunE: rootMain}
			InitFlags(cmd.Flags())
			args := []string{"--file", config, "--discord-embed", "sh", "-c", "exit 0"}
			if fail {
				args[len(args)-1] = "exit 7"
			}
			cmd.SetArgs(args)
			err := cmd.Execute()
			if (err != nil) != fail {
				t.Fatalf("unexpected command result: %v", err)
			}
			want := "Done!"
			if fail {
				want = "exit status 7"
			}
			if title != "sh" || !strings.HasPrefix(description, want+" (") || !strings.HasSuffix(description, ")") {
				t.Fatalf("got %q / %q, want sh / %q with execution time", title, description, want)
			}
			elapsed := strings.TrimSuffix(strings.TrimPrefix(description, want+" ("), ")")
			if _, err := time.ParseDuration(elapsed); err != nil {
				t.Errorf("invalid execution time: %q", elapsed)
			}
		})
	}
}
