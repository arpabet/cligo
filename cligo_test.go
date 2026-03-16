/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"go.arpabet.com/glue"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// captureOutput captures everything written to os.Stdout during f.
func captureOutput(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	r.Close()
	return buf.String()
}

// withArgs temporarily replaces os.Args for the duration of f.
func withArgs(args []string, f func()) {
	old := os.Args
	os.Args = args
	defer func() { os.Args = old }()
	f()
}

// ─── fixtures ────────────────────────────────────────────────────────────────

// shipGroup is a top-level group registered under the "cli" root.
type shipGroup struct {
	Parent CliGroup `cli:"group=cli"`
}

func (g *shipGroup) Group() string          { return "ship" }
func (g *shipGroup) Help() (string, string) { return "Manage ships.", "Manage ships (long)." }

// shipCrewGroup is a sub-group registered under "ship".
type shipCrewGroup struct {
	Parent CliGroup `cli:"group=ship"`
}

func (g *shipCrewGroup) Group() string          { return "crew" }
func (g *shipCrewGroup) Help() (string, string) { return "Manage crew.", "" }

// newShipCmd has a single string positional argument.
type newShipCmd struct {
	Parent CliGroup `cli:"group=ship"`
	Name   string   `cli:"argument=name"`
	ran    bool
}

func (c *newShipCmd) Command() string            { return "new" }
func (c *newShipCmd) Help() (string, string)     { return "Create a ship.", "" }
func (c *newShipCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// setSpeedCmd has a single int positional argument.
type setSpeedCmd struct {
	Parent CliGroup `cli:"group=ship"`
	Speed  int      `cli:"argument=speed"`
	ran    bool
}

func (c *setSpeedCmd) Command() string            { return "setspeed" }
func (c *setSpeedCmd) Help() (string, string)     { return "Set speed.", "" }
func (c *setSpeedCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// moveShipCmd has float positional args and several typed options including a short flag.
type moveShipCmd struct {
	Parent CliGroup `cli:"group=ship"`
	Ship   string   `cli:"argument=ship"`
	X      float64  `cli:"argument=x"`
	Y      float64  `cli:"argument=y"`
	Speed  int      `cli:"option=speed,short=-s,default=10,help=Speed in knots"`
	Dry    bool     `cli:"option=dry,default=false,help=Dry run"`
	Label  string   `cli:"option=label,default=unnamed,help=Label"`
	ran    bool
}

func (c *moveShipCmd) Command() string            { return "move" }
func (c *moveShipCmd) Help() (string, string)     { return "Move a ship.", "" }
func (c *moveShipCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// failCmd always returns an error from Run.
type failCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *failCmd) Command() string            { return "fail" }
func (c *failCmd) Help() (string, string)     { return "Always fails.", "" }
func (c *failCmd) Run(_ context.Context, _ glue.Container) error { return fmt.Errorf("intentional failure") }

// panicErrCmd panics with an error value.
type panicErrCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicErrCmd) Command() string            { return "panicerr" }
func (c *panicErrCmd) Help() (string, string)     { return "Panics with error.", "" }
func (c *panicErrCmd) Run(_ context.Context, _ glue.Container) error { panic(fmt.Errorf("panic error")) }

// panicStrCmd panics with a plain string.
type panicStrCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicStrCmd) Command() string            { return "panicstr" }
func (c *panicStrCmd) Help() (string, string)     { return "Panics with string.", "" }
func (c *panicStrCmd) Run(_ context.Context, _ glue.Container) error { panic("string panic") }

// panicOtherCmd panics with a non-error, non-string value.
type panicOtherCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicOtherCmd) Command() string            { return "panicother" }
func (c *panicOtherCmd) Help() (string, string)     { return "Panics with int.", "" }
func (c *panicOtherCmd) Run(_ context.Context, _ glue.Container) error { panic(42) }

// scopeBean is a DI bean provided by beanCmd's command scope.
type scopeBean struct{ Value string }

// beanCmd implements CliCommandWithBeans, injecting a scopeBean into its scope.
type beanCmd struct {
	Parent CliGroup `cli:"group=ship"`
	ran    bool
}

func (c *beanCmd) Command() string              { return "wbeans" }
func (c *beanCmd) Help() (string, string)       { return "Command with beans.", "" }
func (c *beanCmd) CommandBeans() []interface{}  { return []interface{}{&scopeBean{Value: "injected"}} }
func (c *beanCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// orphanGroup has no CliGroup field, so extractParentGroup returns "".
type orphanGroup struct{}

func (g *orphanGroup) Group() string          { return "orphan" }
func (g *orphanGroup) Help() (string, string) { return "Orphan group.", "" }

// orphanCmd has no CliGroup field.
type orphanCmd struct{}

func (c *orphanCmd) Command() string            { return "orphan" }
func (c *orphanCmd) Help() (string, string)     { return "Orphan command.", "" }
func (c *orphanCmd) Run(_ context.Context, _ glue.Container) error { return nil }

// orphanBeanCmd implements CliCommandWithBeans but has no CliGroup parent field.
type orphanBeanCmd struct{}

func (c *orphanBeanCmd) Command() string             { return "orphanbean" }
func (c *orphanBeanCmd) Help() (string, string)      { return "Orphan bean command.", "" }
func (c *orphanBeanCmd) Run(_ context.Context, _ glue.Container) error { return nil }
func (c *orphanBeanCmd) CommandBeans() []interface{} { return nil }

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
	withArgs([]string{"app", "-V"}, func() {
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

// ctxCheckCmd captures the context it receives so tests can inspect it.
type ctxCheckCmd struct {
	Parent CliGroup `cli:"group=cli"`
	gotCtx context.Context
	ran    bool
}

func (c *ctxCheckCmd) Command() string                              { return "ctxcheck" }
func (c *ctxCheckCmd) Help() (string, string)                       { return "Check context.", "" }
func (c *ctxCheckCmd) Run(ctx context.Context, _ glue.Container) error {
	c.gotCtx = ctx
	c.ran = true
	return ctx.Err()
}

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
