/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "testing"

// ─── extractParentGroup ──────────────────────────────────────────────────────

func TestExtractParentGroup_RootGroup(t *testing.T) {
	if got := extractParentGroup(&shipGroup{}); got != "cli" {
		t.Errorf("expected cli, got %q", got)
	}
}

func TestExtractParentGroup_SubGroup(t *testing.T) {
	if got := extractParentGroup(&shipCrewGroup{}); got != "ship" {
		t.Errorf("expected ship, got %q", got)
	}
}

func TestExtractParentGroup_Command(t *testing.T) {
	if got := extractParentGroup(&newShipCmd{}); got != "ship" {
		t.Errorf("expected ship, got %q", got)
	}
}

func TestExtractParentGroup_NoParentField(t *testing.T) {
	if got := extractParentGroup(&orphanGroup{}); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// ─── extractParentInfo ───────────────────────────────────────────────────────

func TestExtractParentInfo_Hidden(t *testing.T) {
	info := extractParentInfo(&hiddenCmd{})
	if !info.hidden {
		t.Error("expected hidden=true")
	}
	if info.group != "cli" {
		t.Errorf("expected group=cli, got %q", info.group)
	}
}

func TestExtractParentInfo_Alias(t *testing.T) {
	info := extractParentInfo(&aliasedCmd{})
	if info.alias != "n" {
		t.Errorf("expected alias=n, got %q", info.alias)
	}
}

func TestExtractParentInfo_Plain(t *testing.T) {
	info := extractParentInfo(&newShipCmd{})
	if info.hidden {
		t.Error("expected hidden=false")
	}
	if info.alias != "" {
		t.Errorf("expected empty alias, got %q", info.alias)
	}
}
