// SPDX-FileCopyrightText: 2015-2026 Noti Contributors
//
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"codeberg.org/roble/noti/internal/command"
)

func main() {
	command.InitFlags(command.Root.Flags())
	if err := command.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "noti:", err)
		os.Exit(exitStatus(err))
	}
}

func exitStatus(err error) int {
	if err == nil {
		return 0
	}

	var e *exec.ExitError
	if errors.As(err, &e) {
		return e.ExitCode()
	}

	return 1
}
