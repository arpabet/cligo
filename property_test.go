/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"reflect"
	"sort"
	"testing"
)

// ─── parseGlobalProperties ───────────────────────────────────────────────────

func TestParseGlobalProperties_AllForms(t *testing.T) {
	args := []string{
		"-Dhttp-server.bind-address=127.0.0.1:9123", // attached short
		"-D", "log.level=debug", // separated short
		"--property", "a.b=1", // separated long
		"--property=c.d=2", // attached long
		"run", "positional", // ignored
	}
	got := parseGlobalProperties(args)
	want := map[string]string{
		"http-server.bind-address": "127.0.0.1:9123",
		"log.level":                "debug",
		"a.b":                      "1",
		"c.d":                      "2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseGlobalProperties = %v, want %v", got, want)
	}
}

func TestParseGlobalProperties_ValueWithEquals(t *testing.T) {
	// only the first '=' splits key from value
	got := parseGlobalProperties([]string{"-Dquery=a=b=c"})
	if got["query"] != "a=b=c" {
		t.Fatalf("value with '=': got %q, want %q", got["query"], "a=b=c")
	}
}

func TestParseGlobalProperties_None(t *testing.T) {
	if got := parseGlobalProperties([]string{"run", "--verbose"}); len(got) != 0 {
		t.Fatalf("expected no properties, got %v", got)
	}
}

func TestParseGlobalProperties_LastWins(t *testing.T) {
	got := parseGlobalProperties([]string{"-Dk=1", "-Dk=2"})
	if got["k"] != "2" {
		t.Fatalf("last occurrence should win: got %q, want 2", got["k"])
	}
}

// ─── globalPropertyArgSkip ───────────────────────────────────────────────────

func TestGlobalPropertyArgSkip(t *testing.T) {
	cases := []struct {
		arg         string
		wantMatched bool
		wantSkip    int
	}{
		{"-Dk=v", true, 1},
		{"--property=k=v", true, 1},
		{"-D", true, 2},
		{"--property", true, 2},
		{"run", false, 0},
		{"-c", false, 0},
	}
	for _, tc := range cases {
		matched, skip := globalPropertyArgSkip([]string{tc.arg})
		if matched != tc.wantMatched || skip != tc.wantSkip {
			t.Errorf("globalPropertyArgSkip(%q) = (%v,%d), want (%v,%d)", tc.arg, matched, skip, tc.wantMatched, tc.wantSkip)
		}
	}
}

// ─── cliPropertyResolver ─────────────────────────────────────────────────────

func TestCliPropertyResolver(t *testing.T) {
	r := &cliPropertyResolver{props: map[string]string{"x.y": "z"}}

	// priority must outrank glue dotenv(300)/env(200)/file+map(100)
	if r.Priority() <= 300 {
		t.Fatalf("priority = %d, want > 300", r.Priority())
	}
	if v, ok := r.GetProperty("x.y"); !ok || v != "z" {
		t.Fatalf("GetProperty(x.y) = (%q,%v), want (z,true)", v, ok)
	}
	if _, ok := r.GetProperty("missing"); ok {
		t.Fatal("GetProperty(missing) should report ok=false")
	}
	keys := r.Keys()
	sort.Strings(keys)
	if !reflect.DeepEqual(keys, []string{"x.y"}) {
		t.Fatalf("Keys() = %v, want [x.y]", keys)
	}
}
