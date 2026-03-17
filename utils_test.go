/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "testing"

// ─── hasVerbose ──────────────────────────────────────────────────────────────

func TestHasVerbose_Empty(t *testing.T) {
	if hasVerbose([]string{}) {
		t.Error("expected false for empty args")
	}
}

func TestHasVerbose_LongFlag(t *testing.T) {
	if !hasVerbose([]string{"--verbose"}) {
		t.Error("expected true for --verbose")
	}
}

func TestHasVerbose_InMiddle(t *testing.T) {
	if !hasVerbose([]string{"cmd", "--verbose", "--other"}) {
		t.Error("expected true when --verbose is among args")
	}
}

func TestHasVerbose_ShortFlagNotRecognized(t *testing.T) {
	// -l was removed as short flag for verbose
	if hasVerbose([]string{"-l"}) {
		t.Error("expected false: -l is not a short flag for --verbose")
	}
}

func TestHasVerbose_UnrelatedFlags(t *testing.T) {
	if hasVerbose([]string{"--help", "-h", "-v", "--version", "-V"}) {
		t.Error("expected false for unrelated flags")
	}
}
