// SPDX-FileCopyrightText: 2015-2026 Noti Contributors
//
// SPDX-License-Identifier: MIT

//go:build !darwin && !windows
// +build !darwin,!windows

package command

import (
	"fmt"

	"github.com/spf13/viper"
	"codeberg.org/roble/noti/service/espeak"
	"codeberg.org/roble/noti/service/freedesktop"
)

func getBanner(title, message string, v *viper.Viper) notification {
	icon := v.GetString("banner.icon")
	if icon == "" {
		icon = "utilities-terminal"
	}
	return &freedesktop.Notification{
		Summary:       title,
		Body:          message,
		ExpireTimeout: 5000,
		AppIcon:       icon,
	}
}

func getSpeech(title, message string, v *viper.Viper) notification {
	return &espeak.Notification{
		Text:      fmt.Sprintf("%s %s", title, message),
		VoiceName: v.GetString("espeak.voiceName"),
	}
}
