/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

func hasVerbose(args []string) bool {
	for _, arg := range args {
		if arg == "--verbose" {
			return true
		}
	}
	return false
}
