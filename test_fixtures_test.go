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
	"testing"
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

// writeTempFile creates a temporary file and returns its path.
func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
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
func (c *newShipCmd) Run(_ context.Context) error { c.ran = true; return nil }

// setSpeedCmd has a single int positional argument.
type setSpeedCmd struct {
	Parent CliGroup `cli:"group=ship"`
	Speed  int      `cli:"argument=speed"`
	ran    bool
}

func (c *setSpeedCmd) Command() string                               { return "setspeed" }
func (c *setSpeedCmd) Help() (string, string)                        { return "Set speed.", "" }
func (c *setSpeedCmd) Run(_ context.Context) error { c.ran = true; return nil }

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
func (c *moveShipCmd) Run(_ context.Context) error { c.ran = true; return nil }

// failCmd always returns an error from Run.
type failCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *failCmd) Command() string        { return "fail" }
func (c *failCmd) Help() (string, string) { return "Always fails.", "" }
func (c *failCmd) Run(_ context.Context) error {
	return fmt.Errorf("intentional failure")
}

// panicErrCmd panics with an error value.
type panicErrCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicErrCmd) Command() string        { return "panicerr" }
func (c *panicErrCmd) Help() (string, string) { return "Panics with error.", "" }
func (c *panicErrCmd) Run(_ context.Context) error {
	panic(fmt.Errorf("panic error"))
}

// panicStrCmd panics with a plain string.
type panicStrCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicStrCmd) Command() string                               { return "panicstr" }
func (c *panicStrCmd) Help() (string, string)                        { return "Panics with string.", "" }
func (c *panicStrCmd) Run(_ context.Context) error { panic("string panic") }

// panicOtherCmd panics with a non-error, non-string value.
type panicOtherCmd struct {
	Parent CliGroup `cli:"group=ship"`
}

func (c *panicOtherCmd) Command() string                               { return "panicother" }
func (c *panicOtherCmd) Help() (string, string)                        { return "Panics with int.", "" }
func (c *panicOtherCmd) Run(_ context.Context) error { panic(42) }

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
func (c *beanCmd) Run(_ context.Context) error { c.ran = true; return nil }

// orphanGroup has no CliGroup field, so extractParentGroup returns "".
type orphanGroup struct{}

func (g *orphanGroup) Group() string          { return "orphan" }
func (g *orphanGroup) Help() (string, string) { return "Orphan group.", "" }

// orphanCmd has no CliGroup field.
type orphanCmd struct{}

func (c *orphanCmd) Command() string                               { return "orphan" }
func (c *orphanCmd) Help() (string, string)                        { return "Orphan command.", "" }
func (c *orphanCmd) Run(_ context.Context) error { return nil }

// orphanBeanCmd implements CliCommandWithBeans but has no CliGroup parent field.
type orphanBeanCmd struct{}

func (c *orphanBeanCmd) Command() string                               { return "orphanbean" }
func (c *orphanBeanCmd) Help() (string, string)                        { return "Orphan bean command.", "" }
func (c *orphanBeanCmd) Run(_ context.Context) error { return nil }
func (c *orphanBeanCmd) CommandBeans() []interface{}                   { return nil }

// ctxCheckCmd captures the context it receives so tests can inspect it.
type ctxCheckCmd struct {
	Parent CliGroup `cli:"group=cli"`
	gotCtx context.Context
	ran    bool
}

func (c *ctxCheckCmd) Command() string        { return "ctxcheck" }
func (c *ctxCheckCmd) Help() (string, string) { return "Check context.", "" }
func (c *ctxCheckCmd) Run(ctx context.Context) error {
	c.gotCtx = ctx
	c.ran = true
	return ctx.Err()
}

// optArgCmd has a required arg and an optional arg with default.
type optArgCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Name   string   `cli:"argument=name"`
	Color  string   `cli:"argument=color,default=blue"`
	ran    bool
}

func (c *optArgCmd) Command() string                               { return "optarg" }
func (c *optArgCmd) Help() (string, string)                        { return "Optional arg test.", "" }
func (c *optArgCmd) Run(_ context.Context) error { c.ran = true; return nil }

// reqArgCmd has an explicitly required arg.
type reqArgCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Name   string   `cli:"argument=name,required"`
	ran    bool
}

func (c *reqArgCmd) Command() string                               { return "reqarg" }
func (c *reqArgCmd) Help() (string, string)                        { return "Required arg test.", "" }
func (c *reqArgCmd) Run(_ context.Context) error { c.ran = true; return nil }

// optIntArgCmd has an optional int arg with default.
type optIntArgCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Count  int      `cli:"argument=count,default=5"`
	ran    bool
}

func (c *optIntArgCmd) Command() string                               { return "optint" }
func (c *optIntArgCmd) Help() (string, string)                        { return "Optional int arg.", "" }
func (c *optIntArgCmd) Run(_ context.Context) error { c.ran = true; return nil }

// envCmd has an option with env var binding.
type envCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Port   int      `cli:"option=port,default=8080,env=TEST_CLI_PORT,help=Port number"`
	Host   string   `cli:"option=host,default=localhost,env=TEST_CLI_HOST,help=Hostname"`
	ran    bool
}

func (c *envCmd) Command() string                               { return "envcmd" }
func (c *envCmd) Help() (string, string)                        { return "Env var test.", "" }
func (c *envCmd) Run(_ context.Context) error { c.ran = true; return nil }

// hiddenCmd is a command that should not appear in help output.
type hiddenCmd struct {
	Parent CliGroup `cli:"group=cli,hidden"`
	ran    bool
}

func (c *hiddenCmd) Command() string                               { return "secret" }
func (c *hiddenCmd) Help() (string, string)                        { return "Secret command.", "" }
func (c *hiddenCmd) Run(_ context.Context) error { c.ran = true; return nil }

// hiddenGroupDef is a group hidden from help.
type hiddenGroupDef struct {
	Parent CliGroup `cli:"group=cli,hidden"`
}

func (g *hiddenGroupDef) Group() string          { return "internal" }
func (g *hiddenGroupDef) Help() (string, string) { return "Internal group.", "" }

// aliasedCmd has an alias "n" for "new".
type aliasedCmd struct {
	Parent CliGroup `cli:"group=ship,alias=n"`
	Name   string   `cli:"argument=name"`
	ran    bool
}

func (c *aliasedCmd) Command() string                               { return "new" }
func (c *aliasedCmd) Help() (string, string)                        { return "Create a ship.", "" }
func (c *aliasedCmd) Run(_ context.Context) error { c.ran = true; return nil }

// aliasedGroup has an alias "s" for "ship".
type aliasedGroup struct {
	Parent CliGroup `cli:"group=cli,alias=s"`
}

func (g *aliasedGroup) Group() string          { return "ship" }
func (g *aliasedGroup) Help() (string, string) { return "Manage ships.", "" }

// propCmd reads a value from glue properties via the value struct tag.
type propCmd struct {
	Parent  CliGroup `cli:"group=cli"`
	Profile string   `value:"app.profile"`
	Port    string   `value:"app.port"`
	ran     bool
}

func (c *propCmd) Command() string                               { return "propcmd" }
func (c *propCmd) Help() (string, string)                        { return "Prop command.", "" }
func (c *propCmd) Run(_ context.Context) error { c.ran = true; return nil }

// profileCmd is only registered when a profile is active via glue.IfProfile.
type profileCmd struct {
	Parent CliGroup `cli:"group=cli"`
	ran    bool
}

func (c *profileCmd) Command() string                               { return "profcmd" }
func (c *profileCmd) Help() (string, string)                        { return "Profile command.", "" }
func (c *profileCmd) Run(_ context.Context) error { c.ran = true; return nil }

// sliceCmd has slice-typed options.
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
func (c *sliceCmd) Run(_ context.Context) error { c.ran = true; return nil }

// sliceEnvCmd has slice options with env var binding.
type sliceEnvCmd struct {
	Parent CliGroup `cli:"group=cli"`
	Tags   []string `cli:"option=tag,env=APP_TAGS,help=Add a tag"`
	Ports  []int    `cli:"option=port,env=APP_PORTS,help=Add a port"`
	ran    bool
}

func (c *sliceEnvCmd) Command() string                               { return "sliceenvcmd" }
func (c *sliceEnvCmd) Help() (string, string)                        { return "Slice env command.", "" }
func (c *sliceEnvCmd) Run(_ context.Context) error { c.ran = true; return nil }
