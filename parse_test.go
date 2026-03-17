/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "testing"

// ─── parseCliTag ─────────────────────────────────────────────────────────────

func TestParseCliTag_Empty(t *testing.T) {
	result := parseCliTag("")
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestParseCliTag_Single(t *testing.T) {
	result := parseCliTag("argument=name")
	if result["argument"] != "name" {
		t.Errorf("expected argument=name, got %v", result)
	}
}

func TestParseCliTag_Multiple(t *testing.T) {
	result := parseCliTag("option=speed,default=10,help=Speed in knots")
	if result["option"] != "speed" {
		t.Errorf("option: expected speed, got %q", result["option"])
	}
	if result["default"] != "10" {
		t.Errorf("default: expected 10, got %q", result["default"])
	}
	if result["help"] != "Speed in knots" {
		t.Errorf("help: expected 'Speed in knots', got %q", result["help"])
	}
}

func TestParseCliTag_BooleanKey(t *testing.T) {
	result := parseCliTag("required")
	if result["required"] != "true" {
		t.Errorf("expected required=true, got %v", result)
	}
}

func TestParseCliTag_ShortFlag(t *testing.T) {
	result := parseCliTag("option=speed,short=-s")
	if result["short"] != "-s" {
		t.Errorf("expected short=-s, got %q", result["short"])
	}
}

func TestParseCliTag_GroupRef(t *testing.T) {
	result := parseCliTag("group=ship")
	if result["group"] != "ship" {
		t.Errorf("expected group=ship, got %v", result)
	}
}
