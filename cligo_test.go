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
	"path/filepath"
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

func (c *newShipCmd) Command() string                               { return "new" }
func (c *newShipCmd) Help() (string, string)                        { return "Create a ship.", "" }
func (c *newShipCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// setSpeedCmd has a single int positional argument.
type setSpeedCmd struct {
	Parent CliGroup `cli:"group=ship"`
	Speed  int      `cli:"argument=speed"`
	ran    bool
}

func (c *setSpeedCmd) Command() string                               { return "setspeed" }
func (c *setSpeedCmd) Help() (string, string)                        { return "Set speed.", "" }
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

func (c *moveShipCmd) Command() string                               { return "move" }
func (c *moveShipCmd) Help() (string, string)                        { return "Move a ship.", "" }
func (c *moveShipCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// failCmd always returns an error from Run.
type failCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *failCmd) Command() string        { return "fail" }
func (c *failCmd) Help() (string, string) { return "Always fails.", "" }
func (c *failCmd) Run(_ context.Context, _ glue.Container) error {
	return fmt.Errorf("intentional failure")
}

// panicErrCmd panics with an error value.
type panicErrCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicErrCmd) Command() string        { return "panicerr" }
func (c *panicErrCmd) Help() (string, string) { return "Panics with error.", "" }
func (c *panicErrCmd) Run(_ context.Context, _ glue.Container) error {
	panic(fmt.Errorf("panic error"))
}

// panicStrCmd panics with a plain string.
type panicStrCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicStrCmd) Command() string                               { return "panicstr" }
func (c *panicStrCmd) Help() (string, string)                        { return "Panics with string.", "" }
func (c *panicStrCmd) Run(_ context.Context, _ glue.Container) error { panic("string panic") }

// panicOtherCmd panics with a non-error, non-string value.
type panicOtherCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicOtherCmd) Command() string                               { return "panicother" }
func (c *panicOtherCmd) Help() (string, string)                        { return "Panics with int.", "" }
func (c *panicOtherCmd) Run(_ context.Context, _ glue.Container) error { panic(42) }

// scopeBean is a DI bean provided by beanCmd's command scope.
type scopeBean struct{ Value string }

// beanCmd implements CliCommandWithBeans, injecting a scopeBean into its scope.
type beanCmd struct {
	Parent CliGroup `cli:"group=ship"`
	ran    bool
}

func (c *beanCmd) Command() string                               { return "wbeans" }
func (c *beanCmd) Help() (string, string)                        { return "Command with beans.", "" }
func (c *beanCmd) CommandBeans() []interface{}                   { return []interface{}{&scopeBean{Value: "injected"}} }
func (c *beanCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// orphanGroup has no CliGroup field, so extractParentGroup returns "".
type orphanGroup struct{}

func (g *orphanGroup) Group() string          { return "orphan" }
func (g *orphanGroup) Help() (string, string) { return "Orphan group.", "" }

// orphanCmd has no CliGroup field.
type orphanCmd struct{}

func (c *orphanCmd) Command() string                               { return "orphan" }
func (c *orphanCmd) Help() (string, string)                        { return "Orphan command.", "" }
func (c *orphanCmd) Run(_ context.Context, _ glue.Container) error { return nil }

// orphanBeanCmd implements CliCommandWithBeans but has no CliGroup parent field.
type orphanBeanCmd struct{}

func (c *orphanBeanCmd) Command() string                               { return "orphanbean" }
func (c *orphanBeanCmd) Help() (string, string)                        { return "Orphan bean command.", "" }
func (c *orphanBeanCmd) Run(_ context.Context, _ glue.Container) error { return nil }
func (c *orphanBeanCmd) CommandBeans() []interface{}                   { return nil }

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

func (c *ctxCheckCmd) Command() string        { return "ctxcheck" }
func (c *ctxCheckCmd) Help() (string, string) { return "Check context.", "" }
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

// ─── Run: required/optional arguments ─────────────────────────────────────────

// optArgCmd has a required arg and an optional arg with default.
type optArgCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Name   string   `cli:"argument=name"`
	Color  string   `cli:"argument=color,default=blue"`
	ran    bool
}

func (c *optArgCmd) Command() string                               { return "optarg" }
func (c *optArgCmd) Help() (string, string)                        { return "Optional arg test.", "" }
func (c *optArgCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// reqArgCmd has an explicitly required arg.
type reqArgCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Name   string   `cli:"argument=name,required"`
	ran    bool
}

func (c *reqArgCmd) Command() string                               { return "reqarg" }
func (c *reqArgCmd) Help() (string, string)                        { return "Required arg test.", "" }
func (c *reqArgCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// optIntArgCmd has an optional int arg with default.
type optIntArgCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Count  int      `cli:"argument=count,default=5"`
	ran    bool
}

func (c *optIntArgCmd) Command() string                               { return "optint" }
func (c *optIntArgCmd) Help() (string, string)                        { return "Optional int arg.", "" }
func (c *optIntArgCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

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

// ─── Run: environment variable binding ───────────────────────────────────────

// envCmd has an option with env var binding.
type envCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Port   int      `cli:"option=port,default=8080,env=TEST_CLI_PORT,help=Port number"`
	Host   string   `cli:"option=host,default=localhost,env=TEST_CLI_HOST,help=Hostname"`
	ran    bool
}

func (c *envCmd) Command() string                               { return "envcmd" }
func (c *envCmd) Help() (string, string)                        { return "Env var test.", "" }
func (c *envCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

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

// ─── Hidden commands ─────────────────────────────────────────────────────────

// hiddenCmd is a command that should not appear in help output.
type hiddenCmd struct {
	Parent CliGroup `cli:"group=cli,hidden"`
	ran    bool
}

func (c *hiddenCmd) Command() string                               { return "secret" }
func (c *hiddenCmd) Help() (string, string)                        { return "Secret command.", "" }
func (c *hiddenCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// hiddenGroupDef is a group hidden from help.
type hiddenGroupDef struct {
	Parent CliGroup `cli:"group=cli,hidden"`
}

func (g *hiddenGroupDef) Group() string          { return "internal" }
func (g *hiddenGroupDef) Help() (string, string) { return "Internal group.", "" }

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

// ─── Command aliases ─────────────────────────────────────────────────────────

// aliasedCmd has an alias "n" for "new".
type aliasedCmd struct {
	Parent CliGroup `cli:"group=ship,alias=n"`
	Name   string   `cli:"argument=name"`
	ran    bool
}

func (c *aliasedCmd) Command() string                               { return "new" }
func (c *aliasedCmd) Help() (string, string)                        { return "Create a ship.", "" }
func (c *aliasedCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

// aliasedGroup has an alias "s" for "ship".
type aliasedGroup struct {
	Parent CliGroup `cli:"group=cli,alias=s"`
}

func (g *aliasedGroup) Group() string          { return "ship" }
func (g *aliasedGroup) Help() (string, string) { return "Manage ships.", "" }

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

// ─── Config file loading ─────────────────────────────────────────────────────

// propCmd reads a value from glue properties via the value struct tag.
type propCmd struct {
	Parent  CliGroup `cli:"group=cli"`
	Profile string   `value:"app.profile"`
	Port    string   `value:"app.port"`
	ran     bool
}

func (c *propCmd) Command() string                               { return "propcmd" }
func (c *propCmd) Help() (string, string)                        { return "Prop command.", "" }
func (c *propCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestConfigFile_PropertiesFormat(t *testing.T) {
	path := writeTempFile(t, "config.properties", `
# Java properties format
app.profile = production
app.port = 443
`)
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "production" {
		t.Errorf("expected Profile=production, got %q", cmd.Profile)
	}
}

func TestConfigFile_YAMLFormat(t *testing.T) {
	path := writeTempFile(t, "config.yaml", `
app:
  profile: dev
  port: "8080"
`)
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "dev" {
		t.Errorf("expected Profile=dev, got %q", cmd.Profile)
	}
	if cmd.Port != "8080" {
		t.Errorf("expected Port=8080, got %q", cmd.Port)
	}
}

func TestConfigFile_YAMLFormat_YmlExtension(t *testing.T) {
	path := writeTempFile(t, "config.yml", `
app:
  profile: test
  port: "5000"
`)
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "test" {
		t.Errorf("expected Profile=test, got %q", cmd.Profile)
	}
}

func TestConfigFile_JSONFormat(t *testing.T) {
	path := writeTempFile(t, "config.json", `{
  "app": {
    "profile": "live",
    "port": "443"
  }
}`)
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "live" {
		t.Errorf("expected Profile=live, got %q", cmd.Profile)
	}
	if cmd.Port != "443" {
		t.Errorf("expected Port=443, got %q", cmd.Port)
	}
}

func TestConfigFile_FirstExistingFileWins(t *testing.T) {
	path := writeTempFile(t, "app.properties", "app.profile=from-props\napp.port=80")
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile("/nonexistent/app.properties"), ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "from-props" {
		t.Errorf("expected Profile=from-props, got %q", cmd.Profile)
	}
}

func TestConfigFile_NoFileFound_NotAnError(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		captureOutput(func() {
			err := Run(ConfigFile("/nonexistent/a.properties"), ConfigFile("/nonexistent/b.yaml"))
			if err != nil {
				t.Errorf("missing config files should not be an error, got: %v", err)
			}
		})
	})
}

func TestConfigFile_MergesWithExistingProperties(t *testing.T) {
	path := writeTempFile(t, "app.properties", `app.port=7070`)
	props := glue.NewProperties()
	props.Set("app.profile", "preset")

	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(Properties(props), ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "preset" {
		t.Errorf("expected Profile=preset (from existing props), got %q", cmd.Profile)
	}
	if cmd.Port != "7070" {
		t.Errorf("expected Port=7070 (from config file), got %q", cmd.Port)
	}
}

func TestConfigFile_UnsupportedExtension_ReturnsError(t *testing.T) {
	path := writeTempFile(t, "config.xml", `<config/>`)
	withArgs([]string{"app", "--help"}, func() {
		captureOutput(func() {
			err := Run(ConfigFile(path))
			if err == nil {
				t.Error("expected error for unsupported config format")
			}
			if !strings.Contains(err.Error(), "unsupported") {
				t.Errorf("expected 'unsupported' in error, got: %v", err)
			}
		})
	})
}

// ─── profile tests ───────────────────────────────────────────────────────────

// profileCmd is only registered when the "dev" profile is active via glue.IfProfile.
type profileCmd struct {
	Parent CliGroup `cli:"group=cli"`
	ran    bool
}

func (c *profileCmd) Command() string                               { return "profcmd" }
func (c *profileCmd) Help() (string, string)                        { return "Profile command.", "" }
func (c *profileCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

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
		t.Fatal("expected profcmd to run with programmatic Profiles(\"dev\")")
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

// ─── config flag tests ───────────────────────────────────────────────────────

func TestConfigFlag_CLIFlag(t *testing.T) {
	path := writeTempFile(t, "config.properties", "app.profile=fromflag\napp.port=1234")
	cmd := &propCmd{}
	withArgs([]string{"app", "--config", path, "propcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "fromflag" {
		t.Errorf("expected Profile=fromflag, got %q", cmd.Profile)
	}
	if cmd.Port != "1234" {
		t.Errorf("expected Port=1234, got %q", cmd.Port)
	}
}

func TestConfigFlag_CLIFlag_EqualsForm(t *testing.T) {
	path := writeTempFile(t, "config.properties", "app.profile=eqform\napp.port=5678")
	cmd := &propCmd{}
	withArgs([]string{"app", "--config=" + path, "propcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "eqform" {
		t.Errorf("expected Profile=eqform, got %q", cmd.Profile)
	}
}

func TestConfigFlag_Repeated(t *testing.T) {
	// First file exists, second is nonexistent — first wins
	path := writeTempFile(t, "first.properties", "app.profile=first\napp.port=1111")
	cmd := &propCmd{}
	withArgs([]string{"app", "--config", "/nonexistent/second.properties", "--config", path, "propcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "first" {
		t.Errorf("expected Profile=first, got %q", cmd.Profile)
	}
}

func TestConfigFlag_MergesWithOption(t *testing.T) {
	// ConfigFile option provides one path, --config provides another — both are candidates
	path := writeTempFile(t, "flag.properties", "app.profile=flagval\napp.port=9999")
	cmd := &propCmd{}
	withArgs([]string{"app", "--config", path, "propcmd"}, func() {
		if err := Run(ConfigFile("/nonexistent/option.properties"), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "flagval" {
		t.Errorf("expected Profile=flagval, got %q", cmd.Profile)
	}
}

func TestConfigFlag_ShowsInHelp(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run()
		})
		if !strings.Contains(out, "--config") {
			t.Errorf("expected --config in help output, got:\n%s", out)
		}
	})
}

// ─── short flag tests ────────────────────────────────────────────────────────

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

func TestConfigFlag_ShortFlag(t *testing.T) {
	path := writeTempFile(t, "config.properties", "app.profile=short\napp.port=4321")
	cmd := &propCmd{}
	withArgs([]string{"app", "-c", path, "propcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "short" {
		t.Errorf("expected Profile=short, got %q", cmd.Profile)
	}
}

// ─── slice option tests ─────────────────────────────────────────────────────

type sliceCmd struct {
	Parent CliGroup  `cli:"group=cli"`
	Tags   []string  `cli:"option=tag,short=t,help=Add a tag"`
	Ports  []int     `cli:"option=port,help=Add a port"`
	Ratios []float64 `cli:"option=ratio,help=Add a ratio"`
	Flags  []bool    `cli:"option=flag,help=Add a flag"`
	ran    bool
}

func (c *sliceCmd) Command() string                               { return "slicecmd" }
func (c *sliceCmd) Help() (string, string)                        { return "Slice command.", "" }
func (c *sliceCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

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

type sliceEnvCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Tags   []string `cli:"option=tag,env=APP_TAGS,help=Add a tag"`
	Ports  []int    `cli:"option=port,env=APP_PORTS,help=Add a port"`
	ran    bool
}

func (c *sliceEnvCmd) Command() string                               { return "sliceenvcmd" }
func (c *sliceEnvCmd) Help() (string, string)                        { return "Slice env command.", "" }
func (c *sliceEnvCmd) Run(_ context.Context, _ glue.Container) error { c.ran = true; return nil }

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
