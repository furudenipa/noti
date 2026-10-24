// SPDX-FileCopyrightText: 2026 Noti Contributors
//
// SPDX-License-Identifier: MPL-2.0

// Package discord sends notifications using Discord incoming webhooks.
package discord

import (
	"errors"
	"net/http"
	"unicode/utf8"
)

// Notification sends a plain Discord message using content.
type Notification struct {
	Title      string
	Message    string
	Username   string
	AvatarURL  string
	WebhookURL string
	Client     *http.Client
}

// Send validates the message and delivers it through the webhook.
func (n *Notification) Send() error {
	content := n.Message
	if n.Title != "" {
		content = "**" + n.Title + "**"
		if n.Message != "" {
			content += "\n" + n.Message
		}
	}

	if content == "" {
		return errors.New("discord: missing content")
	}
	if utf8.RuneCountInString(content) > 2000 {
		return errors.New("discord: content exceeds 2000 characters")
	}

	return sendWebhook(n.WebhookURL, n.Client, webhookPayload{
		Content:   content,
		Username:  n.Username,
		AvatarURL: n.AvatarURL,
	})
}
