// SPDX-FileCopyrightText: 2015-2026 Noti Contributors
//
// SPDX-License-Identifier: MIT

package command

import (
	"fmt"

	"github.com/spf13/viper"
	"codeberg.org/roble/noti/service/nsuser"
	"codeberg.org/roble/noti/service/say"
)

func getBanner(title, message string, v *viper.Viper) notification {
	return &nsuser.Notification{
		Title:           title,
		InformativeText: message,
		ContentImage:    v.GetString("banner.icon"),
		SoundName:       v.GetString("nsuser.soundName"),
	}
}

func getSpeech(title, message string, v *viper.Viper) notification {
	return &say.Notification{
		Voice: v.GetString("say.voice"),
		Text:  fmt.Sprintf("%s %s", title, message),
		Rate:  200,
	}
}
