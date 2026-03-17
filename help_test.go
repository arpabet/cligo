/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"strings"
	"testing"
)

// ─── Run: help for groups and commands ───────────────────────────────────────

func TestRun_GroupHelp(t *testing.T) {
	withArgs([]string{"app", "ship", "--help"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &newShipCmd{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected help output for group, got: %q", out)
		}
	})
}

func TestRun_CommandHelp_ShowsUsage(t *testing.T) {
	withArgs([]string{"app", "ship", "new", "--help"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &newShipCmd{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected usage in command help, got: %q", out)
		}
	})
}

func TestRun_CommandHelp_ShowsArguments(t *testing.T) {
	withArgs([]string{"app", "ship", "new", "--help"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &newShipCmd{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		// argument names are uppercased in help output
		if !strings.Contains(out, "NAME") {
			t.Errorf("expected argument NAME in command help, got: %q", out)
		}
	})
}

func TestRun_CommandHelp_ShowsOptions(t *testing.T) {
	withArgs([]string{"app", "ship", "move", "--help"}, func() {
		out := captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &moveShipCmd{})); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "--speed") {
			t.Errorf("expected --speed in options output, got: %q", out)
		}
	})
}

func TestRun_CommandHelp_ShowsOptionalArgBrackets(t *testing.T) {
	withArgs([]string{"app", "optarg", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Beans(&optArgCmd{}))
		})
		if !strings.Contains(out, "[COLOR]") {
			t.Errorf("expected [COLOR] for optional arg in usage, got: %q", out)
		}
	})
}

// ─── Hidden commands ─────────────────────────────────────────────────────────

func TestRun_HiddenCommand_NotInHelp(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Beans(&hiddenCmd{}))
		})
		if strings.Contains(out, "secret") {
			t.Errorf("hidden command should not appear in help, got: %q", out)
		}
	})
}

func TestRun_HiddenCommand_StillExecutable(t *testing.T) {
	cmd := &hiddenCmd{}
	withArgs([]string{"app", "secret"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("hidden command should still be executable")
	}
}

func TestRun_HiddenGroup_NotInHelp(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Beans(&hiddenGroupDef{}))
		})
		if strings.Contains(out, "internal") {
			t.Errorf("hidden group should not appear in help, got: %q", out)
		}
	})
}

// ─── Aliases in help ─────────────────────────────────────────────────────────

func TestRun_Alias_ShownInHelp(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Beans(&aliasedGroup{}, &aliasedCmd{}))
		})
		if !strings.Contains(out, "(s)") {
			t.Errorf("expected alias (s) in help output, got: %q", out)
		}
	})
}

func TestRun_CommandAlias_ShownInGroupHelp(t *testing.T) {
	withArgs([]string{"app", "ship", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Beans(&aliasedGroup{}, &aliasedCmd{}))
		})
		if !strings.Contains(out, "(n)") {
			t.Errorf("expected alias (n) in group help output, got: %q", out)
		}
	})
}

// ─── Colored output ──────────────────────────────────────────────────────────

func TestRun_ColorForced_ContainsAnsiCodes(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Color(true), Beans(&shipGroup{}, &newShipCmd{}))
		})
		if !strings.Contains(out, "\033[") {
			t.Errorf("expected ANSI escape codes with Color(true), got: %q", out)
		}
	})
}

func TestRun_ColorDisabled_NoAnsiCodes(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Color(false), Beans(&shipGroup{}, &newShipCmd{}))
		})
		if strings.Contains(out, "\033[") {
			t.Errorf("expected no ANSI codes with Color(false), got: %q", out)
		}
	})
}

func TestRun_ColorForced_CommandHelp_HasAnsiCodes(t *testing.T) {
	withArgs([]string{"app", "ship", "move", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Color(true), Beans(&shipGroup{}, &moveShipCmd{}))
		})
		if !strings.Contains(out, "\033[") {
			t.Errorf("expected ANSI codes in command help with Color(true), got: %q", out)
		}
	})
}

// ─── Env var in help ─────────────────────────────────────────────────────────

func TestRun_EnvVar_ShownInHelp(t *testing.T) {
	withArgs([]string{"app", "envcmd", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Beans(&envCmd{}))
		})
		if !strings.Contains(out, "TEST_CLI_PORT") {
			t.Errorf("expected TEST_CLI_PORT in help output, got: %q", out)
		}
	})
}

// ─── Slice option in help ────────────────────────────────────────────────────

func TestSliceOption_ShowsInHelp(t *testing.T) {
	withArgs([]string{"app", "slicecmd", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run(Beans(&sliceCmd{}))
		})
		if !strings.Contains(out, "--tag") {
			t.Errorf("expected --tag in help output, got:\n%s", out)
		}
	})
}
