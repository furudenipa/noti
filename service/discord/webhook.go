// SPDX-FileCopyrightText: 2026 Noti Contributors
//
// SPDX-License-Identifier: MPL-2.0

package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type webhookPayload struct {
	Content         string  `json:"content,omitempty"`
	Embeds          []embed `json:"embeds,omitempty"`
	Username        string  `json:"username,omitempty"`
	AvatarURL       string  `json:"avatar_url,omitempty"`
	AllowedMentions struct {
		Parse []string `json:"parse"`
	} `json:"allowed_mentions"`
}

func sendWebhook(webhookURL string, client *http.Client, payload webhookPayload) error {
	if webhookURL == "" {
		return errors.New("discord: missing webhook URL")
	}
	u, err := url.Parse(webhookURL)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil {
		return errors.New("discord: invalid webhook URL")
	}
	// Command output must not trigger user, role, or everyone mentions.
	payload.AllowedMentions.Parse = []string{}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: encoding notification: %w", err)
	}

	query := u.Query()
	query.Set("wait", "true")
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return errors.New("discord: invalid webhook request")
	}
	req.Header.Set("Content-Type", "application/json")
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		// The webhook URL includes a secret token; do not include it in errors.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return fmt.Errorf("discord: sending notification: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	var apiErr struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Message != "" {
		return fmt.Errorf("discord: HTTP %d (code %d): %s", resp.StatusCode, apiErr.Code, apiErr.Message)
	}
	return fmt.Errorf("discord: unexpected HTTP status %d", resp.StatusCode)
}
