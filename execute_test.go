/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"context"
	"strings"
	"testing"
)

// ─── Run: global flags ───────────────────────────────────────────────────────

func TestRun_NoArgs_PrintsHelp(t *testing.T) {
	withArgs([]string{"app"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &newShipCmd{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected help output, got: %q", out)
		}
	})
}

func TestRun_HelpShortFlag(t *testing.T) {
	withArgs([]string{"app", "-h"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &newShipCmd{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected help output, got: %q", out)
		}
	})
}

func TestRun_HelpLongFlag(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &newShipCmd{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected help output, got: %q", out)
		}
	})
}

func TestRun_VersionLongFlag(t *testing.T) {
	withArgs([]string{"app", "--version"}, func() {
		out := captureOutput(func() {
			if err := Run(Version("1.0"), Beans(&shipGroup{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "1.0") {
			t.Errorf("expected version in output, got: %q", out)
		}
	})
}

func TestRun_VersionShortFlag(t *testing.T) {
	withArgs([]string{"app", "-v"}, func() {
		out := captureOutput(func() {
			if err := Run(Version("2.5"), Beans(&shipGroup{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "2.5") {
			t.Errorf("expected version in output, got: %q", out)
		}
	})
}

func TestRun_VersionWithBuild(t *testing.T) {
	withArgs([]string{"app", "--version"}, func() {
		out := captureOutput(func() {
			if err := Run(Version("1.0"), Build("99"), Beans(&shipGroup{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "1.0") || !strings.Contains(out, "99") {
			t.Errorf("expected version and build in output, got: %q", out)
		}
	})
}

func TestRun_VerboseFlag_ShowsHelp(t *testing.T) {
	withArgs([]string{"app", "--verbose"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected help output, got: %q", out)
		}
	})
}

func TestRun_UnknownCommand_ReturnsError(t *testing.T) {
	withArgs([]string{"app", "unknown"}, func() {
		captureOutput(func() {
			err := Run(Beans(&shipGroup{}))
			if err == nil {
				t.Error("expected error for unknown command")
			}
			if !strings.Contains(err.Error(), "unknown") {
				t.Errorf("expected 'unknown' in error, got: %v", err)
			}
		})
	})
}

// ─── Run: sub-group navigation ───────────────────────────────────────────────

func TestRun_SubGroup_NoArgs_PrintsHelp(t *testing.T) {
	withArgs([]string{"app", "ship", "crew"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &shipCrewGroup{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected help for sub-group, got: %q", out)
		}
	})
}

func TestRun_SubGroup_Help(t *testing.T) {
	withArgs([]string{"app", "ship", "crew", "--help"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &shipCrewGroup{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected help output for sub-group, got: %q", out)
		}
	})
}

// ─── Run: command execution results ──────────────────────────────────────────

func TestRun_Command_ReturnsError(t *testing.T) {
	withArgs([]string{"app", "ship", "fail"}, func() {
		err := Run(Beans(&shipGroup{}, &failCmd{}))
		if err == nil {
			t.Error("expected error from command")
		}
		if !strings.Contains(err.Error(), "intentional failure") {
			t.Errorf("expected 'intentional failure' in error, got: %v", err)
		}
	})
}

func TestRun_CommandWithBeans_Executed(t *testing.T) {
	cmd := &beanCmd{}
	withArgs([]string{"app", "ship", "wbeans"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("expected command to be executed")
	}
}

// ─── Run: panic recovery ─────────────────────────────────────────────────────

func TestRun_PanicRecovery_Error(t *testing.T) {
	withArgs([]string{"app", "ship", "panicerr"}, func() {
		err := Run(Beans(&shipGroup{}, &panicErrCmd{}))
		if err == nil {
			t.Error("expected error from panic(error)")
		}
		if !strings.Contains(err.Error(), "panic error") {
			t.Errorf("expected 'panic error' in error, got: %v", err)
		}
	})
}

func TestRun_PanicRecovery_String(t *testing.T) {
	withArgs([]string{"app", "ship", "panicstr"}, func() {
		err := Run(Beans(&shipGroup{}, &panicStrCmd{}))
		if err == nil {
			t.Error("expected error from panic(string)")
		}
		if !strings.Contains(err.Error(), "string panic") {
			t.Errorf("expected 'string panic' in error, got: %v", err)
		}
	})
}

func TestRun_PanicRecovery_Other(t *testing.T) {
	withArgs([]string{"app", "ship", "panicother"}, func() {
		err := Run(Beans(&shipGroup{}, &panicOtherCmd{}))
		if err == nil {
			t.Error("expected error from panic(int)")
		}
		if !strings.Contains(err.Error(), "recover:") {
			t.Errorf("expected 'recover:' prefix in error, got: %v", err)
		}
	})
}

// ─── Run: context support ────────────────────────────────────────────────────

func TestRun_DefaultContext_IsNotNil(t *testing.T) {
	cmd := &ctxCheckCmd{}
	withArgs([]string{"app", "ctxcheck"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command was not executed")
	}
	if cmd.gotCtx == nil {
		t.Error("expected non-nil context even without Context option")
	}
}

func TestRun_CustomContext_ThreadedThrough(t *testing.T) {
	type ctxKey struct{}
	parentCtx := context.WithValue(context.Background(), ctxKey{}, "hello")
	cmd := &ctxCheckCmd{}
	withArgs([]string{"app", "ctxcheck"}, func() {
		if err := Run(Context(parentCtx), Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command was not executed")
	}
	if cmd.gotCtx.Value(ctxKey{}) != "hello" {
		t.Error("expected custom context value to be threaded through to command")
	}
}

func TestRun_CancelledContext_Propagated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	cmd := &ctxCheckCmd{}
	withArgs([]string{"app", "ctxcheck"}, func() {
		err := Run(Context(ctx), Beans(cmd))
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command was not executed")
	}
	if cmd.gotCtx.Err() != context.Canceled {
		t.Error("expected context to be cancelled in command")
	}
}

// ─── Run: command aliases ────────────────────────────────────────────────────

func TestRun_CommandAlias_Executes(t *testing.T) {
	cmd := &aliasedCmd{}
	withArgs([]string{"app", "ship", "n", "titanic"}, func() {
		if err := Run(Beans(&aliasedGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command should execute via alias")
	}
	if cmd.Name != "titanic" {
		t.Errorf("expected Name=titanic, got %q", cmd.Name)
	}
}

func TestRun_CommandAlias_PrimaryNameStillWorks(t *testing.T) {
	cmd := &aliasedCmd{}
	withArgs([]string{"app", "ship", "new", "titanic"}, func() {
		if err := Run(Beans(&aliasedGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command should execute via primary name")
	}
}

func TestRun_GroupAlias_Executes(t *testing.T) {
	cmd := &aliasedCmd{}
	withArgs([]string{"app", "s", "new", "titanic"}, func() {
		if err := Run(Beans(&aliasedGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command should execute via group alias")
	}
}
