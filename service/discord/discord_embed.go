// SPDX-FileCopyrightText: 2026 Noti Contributors
//
// SPDX-License-Identifier: MPL-2.0

package discord

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"
)

// EmbedNotification is a Discord embed notification.
type EmbedNotification struct {
	Failed     bool
	Title      string
	Message    string
	Username   string
	AvatarURL  string
	WebhookURL string
	Client     *http.Client
}

type embed struct {
	Color       int    `json:"color"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// Send triggers a Discord embed notification.
func (n *EmbedNotification) Send() error {
	titleLen := utf8.RuneCountInString(strings.TrimSpace(n.Title))
	messageLen := utf8.RuneCountInString(strings.TrimSpace(n.Message))
	if titleLen == 0 && messageLen == 0 {
		return errors.New("discord: empty embed")
	}
	if titleLen > 256 {
		return errors.New("discord: embed title exceeds 256 characters")
	}
	if messageLen > 4096 {
		return errors.New("discord: embed description exceeds 4096 characters")
	}
	color := 0x2ECC71 // Green for success.
	if n.Failed {
		color = 0xE74C3C // Red for failure.
	}
	return sendWebhook(n.WebhookURL, n.Client, webhookPayload{
		Embeds:    []embed{{Title: n.Title, Description: n.Message, Color: color}},
		Username:  n.Username,
		AvatarURL: n.AvatarURL,
	})
}
