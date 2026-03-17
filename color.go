/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "os"

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

func (t *implCliApplication) isColorEnabled() bool {
	if t.color != nil {
		return *t.color
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func (t *implCliApplication) styled(s string, codes ...string) string {
	if !t.isColorEnabled() {
		return s
	}
	var prefix string
	for _, c := range codes {
		prefix += c
	}
	return prefix + s + ansiReset
}
