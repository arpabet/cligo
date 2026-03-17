/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"strings"
	"testing"
)

// ─── Levenshtein ─────────────────────────────────────────────────────────────

func TestLevenshtein_Identical(t *testing.T) {
	if d := levenshtein("abc", "abc"); d != 0 {
		t.Errorf("expected 0, got %d", d)
	}
}

func TestLevenshtein_EmptyStrings(t *testing.T) {
	if d := levenshtein("", "abc"); d != 3 {
		t.Errorf("expected 3, got %d", d)
	}
	if d := levenshtein("abc", ""); d != 3 {
		t.Errorf("expected 3, got %d", d)
	}
}

func TestLevenshtein_SingleEdit(t *testing.T) {
	if d := levenshtein("ship", "shp"); d != 1 {
		t.Errorf("expected 1 (deletion), got %d", d)
	}
	if d := levenshtein("ship", "shiip"); d != 1 {
		t.Errorf("expected 1 (insertion), got %d", d)
	}
	if d := levenshtein("ship", "shep"); d != 1 {
		t.Errorf("expected 1 (substitution), got %d", d)
	}
}

func TestLevenshtein_MultipleEdits(t *testing.T) {
	if d := levenshtein("kitten", "sitting"); d != 3 {
		t.Errorf("expected 3, got %d", d)
	}
}

// ─── "Did you mean?" suggestion ─────────────────────────────────────────────

func TestRun_UnknownCommand_SuggestsClosest(t *testing.T) {
	withArgs([]string{"app", "shp"}, func() {
		captureOutput(func() {
			err := Run(Beans(&shipGroup{}, &newShipCmd{}))
			if err == nil {
				t.Fatal("expected error for unknown command")
			}
			if !strings.Contains(err.Error(), "Did you mean") {
				t.Errorf("expected 'Did you mean' suggestion, got: %v", err)
			}
			if !strings.Contains(err.Error(), "ship") {
				t.Errorf("expected suggestion 'ship', got: %v", err)
			}
		})
	})
}

func TestRun_UnknownCommand_SuggestsCommand(t *testing.T) {
	withArgs([]string{"app", "ship", "nw", "titanic"}, func() {
		captureOutput(func() {
			err := Run(Beans(&shipGroup{}, &newShipCmd{}))
			if err == nil {
				t.Fatal("expected error for unknown command")
			}
			if !strings.Contains(err.Error(), `"new"`) {
				t.Errorf("expected suggestion 'new', got: %v", err)
			}
		})
	})
}

func TestRun_UnknownCommand_NoSuggestionForGarbage(t *testing.T) {
	withArgs([]string{"app", "zzzzzzzzz"}, func() {
		captureOutput(func() {
			err := Run(Beans(&shipGroup{}))
			if err == nil {
				t.Fatal("expected error for unknown command")
			}
			if strings.Contains(err.Error(), "Did you mean") {
				t.Errorf("should not suggest for very different input, got: %v", err)
			}
		})
	})
}

func TestRun_UnknownCommand_SuggestsAlias(t *testing.T) {
	// aliasedGroup has alias "s" for "ship" — typing "x" is close to "s" (distance 1)
	withArgs([]string{"app", "x", "new", "titanic"}, func() {
		captureOutput(func() {
			err := Run(Beans(&aliasedGroup{}, &aliasedCmd{}))
			if err == nil {
				t.Fatal("expected error for unknown command")
			}
			// Should suggest "s" (alias, distance 1 from "x")
			if !strings.Contains(err.Error(), "Did you mean") {
				t.Errorf("expected 'Did you mean' suggestion, got: %v", err)
			}
		})
	})
}
