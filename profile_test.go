/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"strings"
	"testing"

	"go.arpabet.com/glue"
)

// ─── Profile tests ──────────────────────────────────────────────────────────

func TestProfile_CLIFlag_ActivatesProfile(t *testing.T) {
	cmd := &profileCmd{}
	withArgs([]string{"app", "--profile", "dev", "profcmd"}, func() {
		if err := Run(Beans(glue.IfProfile("dev", cmd))); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Fatal("expected profcmd to run when dev profile is active")
	}
}

func TestProfile_CLIFlag_InactiveProfile_CommandNotRegistered(t *testing.T) {
	cmd := &profileCmd{}
	withArgs([]string{"app", "--profile", "staging", "profcmd"}, func() {
		captureOutput(func() {
			err := Run(Beans(glue.IfProfile("dev", cmd)))
			if err == nil {
				t.Fatal("expected error for unregistered command")
			}
			if !strings.Contains(err.Error(), "unknown command") {
				t.Errorf("expected 'unknown command' error, got: %v", err)
			}
		})
	})
	if cmd.ran {
		t.Fatal("profcmd should not run when dev profile is not active")
	}
}

func TestProfile_CLIFlag_CommaSeparated(t *testing.T) {
	devCmd := &profileCmd{}
	withArgs([]string{"app", "--profile", "dev,staging", "profcmd"}, func() {
		if err := Run(Beans(glue.IfProfile("staging", devCmd))); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !devCmd.ran {
		t.Fatal("expected profcmd to run with staging profile from comma-separated list")
	}
}

func TestProfile_CLIFlag_EqualsForm(t *testing.T) {
	cmd := &profileCmd{}
	withArgs([]string{"app", "--profile=dev", "profcmd"}, func() {
		if err := Run(Beans(glue.IfProfile("dev", cmd))); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Fatal("expected profcmd to run with --profile=dev")
	}
}

func TestProfile_Option_Programmatic(t *testing.T) {
	cmd := &profileCmd{}
	withArgs([]string{"app", "profcmd"}, func() {
		if err := Run(Profile("dev"), Beans(glue.IfProfile("dev", cmd))); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Fatal("expected profcmd to run with programmatic Profile(\"dev\")")
	}
}

func TestProfile_MergesCLIAndProgrammatic(t *testing.T) {
	// Use a command gated on "dev&staging" — requires both profiles active.
	cmd := &profileCmd{}
	withArgs([]string{"app", "--profile", "staging", "profcmd"}, func() {
		if err := Run(Profile("dev"), Beans(glue.IfProfile("dev&staging", cmd))); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Fatal("expected profcmd to run when both dev (programmatic) and staging (CLI) profiles are active")
	}
}

func TestProfile_ShowsInHelp(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run()
		})
		if !strings.Contains(out, "--profile") {
			t.Errorf("expected --profile in help output, got:\n%s", out)
		}
	})
}

func TestProfile_NoProfile_SkipsProfileBeans(t *testing.T) {
	cmd := &profileCmd{}
	withArgs([]string{"app", "profcmd"}, func() {
		captureOutput(func() {
			err := Run(Beans(glue.IfProfile("dev", cmd)))
			if err == nil {
				t.Fatal("expected error when no profile is active and command is profile-gated")
			}
		})
	})
	if cmd.ran {
		t.Fatal("profcmd should not run without any active profile")
	}
}

func TestProfile_CLIFlag_Repeated(t *testing.T) {
	cmd := &profileCmd{}
	withArgs([]string{"app", "--profile", "dev", "--profile", "staging", "profcmd"}, func() {
		if err := Run(Beans(glue.IfProfile("dev&staging", cmd))); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Fatal("expected profcmd to run with repeated --profile flags")
	}
}

func TestProfile_ShortFlag(t *testing.T) {
	cmd := &profileCmd{}
	withArgs([]string{"app", "-p", "dev", "profcmd"}, func() {
		if err := Run(Beans(glue.IfProfile("dev", cmd))); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Fatal("expected profcmd to run with -p dev")
	}
}

func TestProfile_ShortFlag_Repeated(t *testing.T) {
	cmd := &profileCmd{}
	withArgs([]string{"app", "-p", "dev", "-p", "staging", "profcmd"}, func() {
		if err := Run(Beans(glue.IfProfile("dev&staging", cmd))); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Fatal("expected profcmd to run with repeated -p flags")
	}
}
