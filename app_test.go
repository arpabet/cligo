/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "testing"

// ─── Echo ────────────────────────────────────────────────────────────────────

func TestEcho_WithFormat(t *testing.T) {
	out := captureOutput(func() { Echo("hello %s", "world") })
	if out != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got %q", out)
	}
}

func TestEcho_PlainText(t *testing.T) {
	out := captureOutput(func() { Echo("plain text") })
	if out != "plain text\n" {
		t.Errorf("expected 'plain text\\n', got %q", out)
	}
}

func TestEcho_MultipleArgs(t *testing.T) {
	out := captureOutput(func() { Echo("%d + %d = %d", 1, 2, 3) })
	if out != "1 + 2 = 3\n" {
		t.Errorf("expected '1 + 2 = 3\\n', got %q", out)
	}
}

// ─── New ─────────────────────────────────────────────────────────────────────

func TestNew_DefaultNameFromArgs(t *testing.T) {
	withArgs([]string{"/path/to/myapp"}, func() {
		app := New()
		if app.Name() != "myapp" {
			t.Errorf("expected myapp, got %q", app.Name())
		}
	})
}

func TestNew_WithOptions(t *testing.T) {
	withArgs([]string{"app"}, func() {
		app := New(
			Name("testapp"),
			Title("Test Application"),
			Help("Test help"),
			Version("2.0"),
			Build("42"),
			Verbose(true),
		)
		if app.Name() != "testapp" {
			t.Errorf("Name: expected testapp, got %q", app.Name())
		}
		if app.Title() != "Test Application" {
			t.Errorf("Title: expected 'Test Application', got %q", app.Title())
		}
		if app.Help() != "Test help" {
			t.Errorf("Help: expected 'Test help', got %q", app.Help())
		}
		if app.Version() != "2.0" {
			t.Errorf("Version: expected 2.0, got %q", app.Version())
		}
		if app.Build() != "42" {
			t.Errorf("Build: expected 42, got %q", app.Build())
		}
		if !app.Verbose() {
			t.Error("Verbose: expected true")
		}
	})
}

func TestNew_VerboseDetectedFromArgs(t *testing.T) {
	withArgs([]string{"app", "--verbose"}, func() {
		app := New()
		if !app.Verbose() {
			t.Error("expected Verbose=true when --verbose is in os.Args")
		}
	})
}

func TestNew_Nope(t *testing.T) {
	withArgs([]string{"app"}, func() {
		app := New(Nope())
		if app == nil {
			t.Error("expected non-nil app from New(Nope())")
		}
	})
}

func TestNew_BeansStored(t *testing.T) {
	withArgs([]string{"app"}, func() {
		extra := &struct{}{}
		app := New(Beans(extra))
		found := false
		for _, b := range app.getBeans() {
			if b == extra {
				found = true
			}
		}
		if !found {
			t.Error("expected custom bean to be in app.getBeans()")
		}
	})
}

// ─── Register* ───────────────────────────────────────────────────────────────

func TestRegisterGroup_Valid(t *testing.T) {
	withArgs([]string{"app"}, func() {
		app := New()
		if err := app.RegisterGroup(&shipGroup{}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRegisterGroup_NoParent(t *testing.T) {
	withArgs([]string{"app"}, func() {
		app := New()
		if err := app.RegisterGroup(&orphanGroup{}); err == nil {
			t.Error("expected error for group with no CliGroup parent field")
		}
	})
}

func TestRegisterCommand_Valid(t *testing.T) {
	withArgs([]string{"app"}, func() {
		app := New()
		_ = app.RegisterGroup(&shipGroup{})
		if err := app.RegisterCommand(&newShipCmd{}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRegisterCommand_NoParent(t *testing.T) {
	withArgs([]string{"app"}, func() {
		app := New()
		if err := app.RegisterCommand(&orphanCmd{}); err == nil {
			t.Error("expected error for command with no CliGroup parent field")
		}
	})
}

func TestRegisterCommandWithBeans_Valid(t *testing.T) {
	withArgs([]string{"app"}, func() {
		app := New()
		_ = app.RegisterGroup(&shipGroup{})
		if err := app.RegisterCommandWithBeans(&beanCmd{}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRegisterCommandWithBeans_NoParent(t *testing.T) {
	withArgs([]string{"app"}, func() {
		app := New()
		if err := app.RegisterCommandWithBeans(&orphanBeanCmd{}); err == nil {
			t.Error("expected error for CliCommandWithBeans with no CliGroup parent field")
		}
	})
}
