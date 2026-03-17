/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"os"
	"strings"
	"testing"
)

// ─── Run: command argument parsing ───────────────────────────────────────────

func TestRun_Command_StringArg(t *testing.T) {
	cmd := &newShipCmd{}
	withArgs([]string{"app", "ship", "new", "titanic"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command was not executed")
	}
	if cmd.Name != "titanic" {
		t.Errorf("expected Name=titanic, got %q", cmd.Name)
	}
}

func TestRun_Command_IntArg(t *testing.T) {
	cmd := &setSpeedCmd{}
	withArgs([]string{"app", "ship", "setspeed", "42"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command was not executed")
	}
	if cmd.Speed != 42 {
		t.Errorf("expected Speed=42, got %d", cmd.Speed)
	}
}

func TestRun_Command_FloatArgs(t *testing.T) {
	cmd := &moveShipCmd{}
	withArgs([]string{"app", "ship", "move", "titanic", "1.5", "2.5"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command was not executed")
	}
	if cmd.Ship != "titanic" {
		t.Errorf("expected Ship=titanic, got %q", cmd.Ship)
	}
	if cmd.X != 1.5 {
		t.Errorf("expected X=1.5, got %v", cmd.X)
	}
	if cmd.Y != 2.5 {
		t.Errorf("expected Y=2.5, got %v", cmd.Y)
	}
}

func TestRun_Command_MissingArg_ReturnsError(t *testing.T) {
	withArgs([]string{"app", "ship", "new"}, func() {
		captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &newShipCmd{})); err == nil {
				t.Error("expected error for missing positional argument")
			}
		})
	})
}

func TestRun_Command_InvalidIntArg_ReturnsError(t *testing.T) {
	withArgs([]string{"app", "ship", "setspeed", "notanint"}, func() {
		captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &setSpeedCmd{})); err == nil {
				t.Error("expected error for invalid integer argument")
			}
		})
	})
}

func TestRun_Command_InvalidFloatArg_ReturnsError(t *testing.T) {
	withArgs([]string{"app", "ship", "move", "titanic", "notafloat", "2.0"}, func() {
		captureOutput(func() {
			if err := Run(Beans(&shipGroup{}, &moveShipCmd{})); err == nil {
				t.Error("expected error for invalid float argument")
			}
		})
	})
}

// ─── Run: command option parsing ─────────────────────────────────────────────

func TestRun_Command_IntOption_LongFlag(t *testing.T) {
	cmd := &moveShipCmd{}
	withArgs([]string{"app", "ship", "move", "titanic", "1.0", "2.0", "--speed=30"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Speed != 30 {
		t.Errorf("expected Speed=30, got %d", cmd.Speed)
	}
}

func TestRun_Command_IntOption_ShortFlag(t *testing.T) {
	cmd := &moveShipCmd{}
	withArgs([]string{"app", "ship", "move", "titanic", "1.0", "2.0", "-s", "25"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Speed != 25 {
		t.Errorf("expected Speed=25, got %d", cmd.Speed)
	}
}

func TestRun_Command_BoolOption(t *testing.T) {
	cmd := &moveShipCmd{}
	withArgs([]string{"app", "ship", "move", "titanic", "1.0", "2.0", "--dry"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.Dry {
		t.Error("expected Dry=true when --dry is passed")
	}
}

func TestRun_Command_StringOption(t *testing.T) {
	cmd := &moveShipCmd{}
	withArgs([]string{"app", "ship", "move", "titanic", "1.0", "2.0", "--label=flagship"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Label != "flagship" {
		t.Errorf("expected Label=flagship, got %q", cmd.Label)
	}
}

func TestRun_Command_DefaultOptionValues(t *testing.T) {
	cmd := &moveShipCmd{}
	withArgs([]string{"app", "ship", "move", "titanic", "1.0", "2.0"}, func() {
		if err := Run(Beans(&shipGroup{}, cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Speed != 10 {
		t.Errorf("expected Speed=10 (default), got %d", cmd.Speed)
	}
	if cmd.Label != "unnamed" {
		t.Errorf("expected Label='unnamed' (default), got %q", cmd.Label)
	}
	if cmd.Dry {
		t.Error("expected Dry=false (default)")
	}
}

// ─── Run: required/optional arguments ─────────────────────────────────────────

func TestRun_OptionalArg_UsesDefault(t *testing.T) {
	cmd := &optArgCmd{}
	withArgs([]string{"app", "optarg", "alice"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !cmd.ran {
		t.Error("command was not executed")
	}
	if cmd.Name != "alice" {
		t.Errorf("expected Name=alice, got %q", cmd.Name)
	}
	if cmd.Color != "blue" {
		t.Errorf("expected Color=blue (default), got %q", cmd.Color)
	}
}

func TestRun_OptionalArg_ProvidedExplicitly(t *testing.T) {
	cmd := &optArgCmd{}
	withArgs([]string{"app", "optarg", "alice", "red"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Color != "red" {
		t.Errorf("expected Color=red, got %q", cmd.Color)
	}
}

func TestRun_RequiredArg_Missing_ReturnsError(t *testing.T) {
	withArgs([]string{"app", "optarg"}, func() {
		captureOutput(func() {
			err := Run(Beans(&optArgCmd{}))
			if err == nil {
				t.Error("expected error for missing required argument")
			}
			if !strings.Contains(err.Error(), "required") {
				t.Errorf("expected 'required' in error, got: %v", err)
			}
		})
	})
}

func TestRun_ExplicitRequiredTag_Missing_ReturnsError(t *testing.T) {
	withArgs([]string{"app", "reqarg"}, func() {
		captureOutput(func() {
			err := Run(Beans(&reqArgCmd{}))
			if err == nil {
				t.Error("expected error for missing explicitly required argument")
			}
		})
	})
}

func TestRun_OptionalIntArg_UsesDefault(t *testing.T) {
	cmd := &optIntArgCmd{}
	withArgs([]string{"app", "optint"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Count != 5 {
		t.Errorf("expected Count=5 (default), got %d", cmd.Count)
	}
}

func TestRun_OptionalIntArg_ProvidedExplicitly(t *testing.T) {
	cmd := &optIntArgCmd{}
	withArgs([]string{"app", "optint", "99"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Count != 99 {
		t.Errorf("expected Count=99, got %d", cmd.Count)
	}
}

// ─── Run: environment variable binding ───────────────────────────────────────

func TestRun_EnvVar_UsedWhenFlagNotSet(t *testing.T) {
	os.Setenv("TEST_CLI_PORT", "9090")
	defer os.Unsetenv("TEST_CLI_PORT")

	cmd := &envCmd{}
	withArgs([]string{"app", "envcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Port != 9090 {
		t.Errorf("expected Port=9090 from env, got %d", cmd.Port)
	}
}

func TestRun_EnvVar_FlagTakesPrecedence(t *testing.T) {
	os.Setenv("TEST_CLI_PORT", "9090")
	defer os.Unsetenv("TEST_CLI_PORT")

	cmd := &envCmd{}
	withArgs([]string{"app", "envcmd", "--port=3000"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Port != 3000 {
		t.Errorf("expected Port=3000 from explicit flag, got %d", cmd.Port)
	}
}

func TestRun_EnvVar_FallsBackToDefault(t *testing.T) {
	os.Unsetenv("TEST_CLI_PORT")

	cmd := &envCmd{}
	withArgs([]string{"app", "envcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Port != 8080 {
		t.Errorf("expected Port=8080 (default), got %d", cmd.Port)
	}
}

func TestRun_EnvVar_StringOption(t *testing.T) {
	os.Setenv("TEST_CLI_HOST", "0.0.0.0")
	defer os.Unsetenv("TEST_CLI_HOST")

	cmd := &envCmd{}
	withArgs([]string{"app", "envcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if cmd.Host != "0.0.0.0" {
		t.Errorf("expected Host=0.0.0.0 from env, got %q", cmd.Host)
	}
}

// ─── Slice options ───────────────────────────────────────────────────────────

func TestSliceOption_StringArray_Repeated(t *testing.T) {
	cmd := &sliceCmd{}
	withArgs([]string{"app", "slicecmd", "--tag=foo", "--tag=bar", "--tag=baz"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if len(cmd.Tags) != 3 || cmd.Tags[0] != "foo" || cmd.Tags[1] != "bar" || cmd.Tags[2] != "baz" {
		t.Errorf("expected Tags=[foo bar baz], got %v", cmd.Tags)
	}
}

func TestSliceOption_StringArray_ShortFlag(t *testing.T) {
	cmd := &sliceCmd{}
	withArgs([]string{"app", "slicecmd", "-t", "alpha", "-t", "beta"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if len(cmd.Tags) != 2 || cmd.Tags[0] != "alpha" || cmd.Tags[1] != "beta" {
		t.Errorf("expected Tags=[alpha beta], got %v", cmd.Tags)
	}
}

func TestSliceOption_IntSlice(t *testing.T) {
	cmd := &sliceCmd{}
	withArgs([]string{"app", "slicecmd", "--port=8080", "--port=9090"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if len(cmd.Ports) != 2 || cmd.Ports[0] != 8080 || cmd.Ports[1] != 9090 {
		t.Errorf("expected Ports=[8080 9090], got %v", cmd.Ports)
	}
}

func TestSliceOption_Float64Slice(t *testing.T) {
	cmd := &sliceCmd{}
	withArgs([]string{"app", "slicecmd", "--ratio=1.5", "--ratio=2.7"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if len(cmd.Ratios) != 2 || cmd.Ratios[0] != 1.5 || cmd.Ratios[1] != 2.7 {
		t.Errorf("expected Ratios=[1.5 2.7], got %v", cmd.Ratios)
	}
}

func TestSliceOption_BoolSlice(t *testing.T) {
	cmd := &sliceCmd{}
	withArgs([]string{"app", "slicecmd", "--flag=true", "--flag=false", "--flag=true"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if len(cmd.Flags) != 3 || cmd.Flags[0] != true || cmd.Flags[1] != false || cmd.Flags[2] != true {
		t.Errorf("expected Flags=[true false true], got %v", cmd.Flags)
	}
}

func TestSliceOption_Empty_ReturnsNil(t *testing.T) {
	cmd := &sliceCmd{}
	withArgs([]string{"app", "slicecmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Tags != nil {
		t.Errorf("expected Tags=nil when not provided, got %v", cmd.Tags)
	}
}

func TestSliceOption_EnvVar_StringArray(t *testing.T) {
	t.Setenv("APP_TAGS", "x,y,z")
	cmd := &sliceEnvCmd{}
	withArgs([]string{"app", "sliceenvcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if len(cmd.Tags) != 3 || cmd.Tags[0] != "x" || cmd.Tags[1] != "y" || cmd.Tags[2] != "z" {
		t.Errorf("expected Tags=[x y z], got %v", cmd.Tags)
	}
}

func TestSliceOption_EnvVar_IntSlice(t *testing.T) {
	t.Setenv("APP_PORTS", "80,443,8080")
	cmd := &sliceEnvCmd{}
	withArgs([]string{"app", "sliceenvcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if len(cmd.Ports) != 3 || cmd.Ports[0] != 80 || cmd.Ports[1] != 443 || cmd.Ports[2] != 8080 {
		t.Errorf("expected Ports=[80 443 8080], got %v", cmd.Ports)
	}
}

func TestSliceOption_CLIOverridesEnvVar(t *testing.T) {
	t.Setenv("APP_TAGS", "from-env")
	cmd := &sliceEnvCmd{}
	withArgs([]string{"app", "sliceenvcmd", "--tag=from-cli"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if len(cmd.Tags) != 1 || cmd.Tags[0] != "from-cli" {
		t.Errorf("expected Tags=[from-cli] (CLI overrides env), got %v", cmd.Tags)
	}
}
