# cligo

A declarative CLI framework for Go, inspired by Python's [Click](https://click.palletsprojects.com/) library. Built on top of the [glue](https://go.arpabet.com/glue) dependency injection framework.

Define commands as Go structs, declare arguments and options via struct tags, and let cligo handle parsing, help generation, and dependency injection automatically.

## Installation

```
go get go.arpabet.com/cligo
```

## Quick Start

```go
package main

import (
    "context"

    "go.arpabet.com/cligo"
    "go.arpabet.com/glue"
)

type Greet struct {
    Parent cligo.CliGroup `cli:"group=cli"`
    Name   string         `cli:"argument=name"`
}

func (cmd *Greet) Command() string                                  { return "greet" }
func (cmd *Greet) Help() (string, string)                           { return "Greet someone.", "" }
func (cmd *Greet) Run(ctx context.Context, c glue.Container) error  {
    cligo.Echo("Hello, %s!", cmd.Name)
    return nil
}

func main() {
    cligo.Main(
        cligo.Name("hello"),
        cligo.Version("1.0.0"),
        cligo.Help("A friendly greeter."),
        cligo.Beans(&Greet{}),
    )
}
```

```
$ hello greet World
Hello, World!
$ hello --version
hello Version 1.0.0
$ hello --help
Usage: hello  [OPTIONS] COMMAND [ARGS]...

A friendly greeter.

Options:
  --version  Show the version and exit.
  --verbose  Show extended logging information.
  --help     Show this message and exit.

Commands:
  greet	Greet someone.
```

## Core Concepts

### Groups

Groups organize commands into a hierarchy (like `git remote add`). Implement the `CliGroup` interface and embed a `cligo.CliGroup` field to declare the parent.

Root-level groups attach to the built-in root `"cli"`:

```go
type Ship struct {
    Parent cligo.CliGroup `cli:"group=cli"`
}

func (g *Ship) Group() string              { return "ship" }
func (g *Ship) Help() (string, string)     { return "Manage ships.", "" }
```

Sub-groups reference their parent by name:

```go
type ShipCrew struct {
    Parent cligo.CliGroup `cli:"group=ship"`
}

func (g *ShipCrew) Group() string          { return "crew" }
func (g *ShipCrew) Help() (string, string) { return "Manage ship crew.", "" }
```

This produces the hierarchy: `app ship crew <command>`.

### Commands

Commands are the leaf nodes that execute logic. Implement the `CliCommand` interface:

```go
type ShipNew struct {
    Parent cligo.CliGroup `cli:"group=ship"`
    Name   string         `cli:"argument=name"`
}

func (cmd *ShipNew) Command() string                                  { return "new" }
func (cmd *ShipNew) Help() (string, string)                           { return "Create a new ship.", "" }
func (cmd *ShipNew) Run(ctx context.Context, c glue.Container) error  {
    cligo.Echo("Created ship %s", cmd.Name)
    return nil
}
```

The `Help()` method returns `(shortDescription, optionalLongDescription)`. The short description appears in command listings; the long description appears in the command's own `--help` output. If the long description is empty, the short description is used for both.

### Arguments

Positional arguments are declared with `cli:"argument=<name>"` struct tags. They are parsed in the order they appear in the struct:

```go
type Move struct {
    Parent cligo.CliGroup `cli:"group=ship"`
    Ship   string         `cli:"argument=ship"`
    X      float64        `cli:"argument=x"`
    Y      float64        `cli:"argument=y"`
}
```

```
$ app ship move titanic 1.5 2.5
```

Supported argument types: `string`, `int` (all sizes), `float32`, `float64`.

### Options

Named flags are declared with `cli:"option=<name>"` and support defaults, help text, and short flags:

```go
type Move struct {
    Parent cligo.CliGroup `cli:"group=ship"`
    Ship   string         `cli:"argument=ship"`
    Speed  int            `cli:"option=speed,short=-s,default=10,help=Speed in knots"`
    Dry    bool           `cli:"option=dry,default=false,help=Dry run mode"`
    Label  string         `cli:"option=label,default=unnamed,help=Ship label"`
}
```

```
$ app ship move titanic --speed=20
$ app ship move titanic -s 20
$ app ship move titanic --dry --label=flagship
```

Supported option types: `string`, `int` (all sizes), `float32`, `float64`, `bool`.

## Struct Tag Reference

All metadata is declared in the `cli` struct tag with comma-separated `key=value` pairs:

| Tag | Description | Example |
|-----|-------------|---------|
| `group=<name>` | Parent group (required on the `CliGroup` field) | `cli:"group=cli"` |
| `argument=<name>` | Positional argument | `cli:"argument=name"` |
| `option=<name>` | Named flag/option | `cli:"option=speed"` |
| `short=-<char>` | Single-character short flag | `cli:"option=speed,short=-s"` |
| `default=<value>` | Default value for an option | `cli:"option=speed,default=10"` |
| `help=<text>` | Help text for an option | `cli:"option=speed,help=Speed in knots"` |

Tags can be combined: `cli:"option=speed,short=-s,default=10,help=Speed in knots"`

## Application Options

Configure the application using functional options passed to `Main()` or `Run()`:

```go
cligo.Main(
    cligo.Name("myapp"),          // Binary name (defaults to os.Args[0])
    cligo.Title("My Application"),// Display title shown in --version
    cligo.Help("Description."),   // Help text shown in usage
    cligo.Version("1.0.0"),       // Enables --version / -V flag
    cligo.Build("abc123"),        // Build identifier shown alongside version
    cligo.Verbose(true),          // Force verbose mode
    cligo.Context(ctx),           // Custom context (defaults to signal-aware context)
    cligo.Beans(beans...),        // Register groups and commands
    cligo.Properties(props),      // Glue properties for DI
)
```

| Option | Description |
|--------|-------------|
| `Name(s)` | Application name (defaults to binary name) |
| `Title(s)` | Display title for version output |
| `Help(s)` | Description shown in help output |
| `Version(s)` | Version string; enables `--version` / `-V` |
| `Build(s)` | Build identifier shown with version |
| `Verbose(b)` | Force verbose mode on |
| `Context(ctx)` | Custom `context.Context` (defaults to signal-aware context) |
| `Beans(b...)` | Groups, commands, and other DI beans |
| `Properties(p)` | Glue properties for dependency injection |
| `Nope()` | No-op (useful for conditional options) |

## Global Flags

These flags are handled automatically:

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help for the application, group, or command |
| `--version`, `-V` | Show version and build info (requires `Version()` option) |
| `--verbose` | Enable verbose logging via glue |

## Context & Signal Handling

Every command receives a `context.Context` as the first argument to `Run()`. By default, cligo creates a signal-aware context that is cancelled on `SIGINT` or `SIGTERM`, enabling graceful shutdown:

```go
func (cmd *Serve) Run(ctx context.Context, c glue.Container) error {
    server := &http.Server{Addr: ":8080"}
    go func() {
        <-ctx.Done() // cancelled on Ctrl+C
        server.Shutdown(context.Background())
    }()
    return server.ListenAndServe()
}
```

To provide a custom context (e.g., with a timeout or custom values), use the `Context()` option:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

cligo.Main(
    cligo.Context(ctx),
    cligo.Beans(&Serve{}),
)
```

If no `Context()` option is provided, cligo creates a context via `signal.NotifyContext` that listens for `SIGINT` and `SIGTERM`.

## Dependency Injection

Cligo is built on top of [glue](https://go.arpabet.com/glue), a dependency injection framework. Every command receives a `context.Context` and a `glue.Container` in its `Run()` method, giving access to cancellation signals and all registered beans.

All structs passed via `Beans()` are automatically registered in the glue container. Groups and commands are discovered by interface type and registered with the CLI application.

### Command-Scoped Beans

Commands that need private dependencies can implement `CliCommandWithBeans`:

```go
type DatabaseService struct { /* ... */ }

type MigrateCmd struct {
    Parent cligo.CliGroup `cli:"group=cli"`
}

func (cmd *MigrateCmd) Command() string                                  { return "migrate" }
func (cmd *MigrateCmd) Help() (string, string)                           { return "Run migrations.", "" }
func (cmd *MigrateCmd) CommandBeans() []interface{}  {
    return []interface{}{&DatabaseService{}}
}
func (cmd *MigrateCmd) Run(ctx context.Context, c glue.Container) error  {
    // c is an extended container with DatabaseService available
    // ctx carries cancellation/timeout signals
    return nil
}
```

The command-scoped beans are added to a child container that is created before `Run()` is called and closed after it returns.

### Property Injection

Use glue properties to inject configuration values into command fields:

```go
type AddUser struct {
    Parent  cligo.CliGroup `cli:"group=users"`
    Name    string         `cli:"argument=name"`
    Profile string         `value:"profiles.active"` // injected by glue
}

func main() {
    properties := glue.NewProperties()
    properties.Set("profiles.active", "dev")

    cligo.Main(
        cligo.Properties(properties),
        cligo.Beans(&AddUser{}),
    )
}
```

## Entry Points

| Function | Description |
|----------|-------------|
| `cligo.Main(opts...)` | Parse args, run the matched command, call `os.Exit(1)` on error |
| `cligo.Run(opts...)` | Same as `Main` but returns the error instead of exiting |
| `cligo.New(opts...)` | Create a `CliApplication` without executing (for advanced use) |

## Interfaces

```go
// CliGroup defines a command group (sub-command namespace).
type CliGroup interface {
    Group() string
    Help() (short string, optionalLong string)
}

// CliCommand defines an executable command.
type CliCommand interface {
    Command() string
    Help() (short string, optionalLong string)
    Run(ctx context.Context, c glue.Container) error
}

// CliCommandWithBeans extends CliCommand with command-scoped DI beans.
type CliCommandWithBeans interface {
    CliCommand
    CommandBeans() []interface{}
}
```

## Examples

See the [examples/](examples/) directory:

- **[basic](examples/basic/)** -- Flat commands (no groups) with arguments and options
- **[naval](examples/naval/)** -- Nested groups (`ship`, `mine`) with multiple commands and short flags
- **[props](examples/props/)** -- Property injection via glue

Run an example:

```
go run ./examples/naval ship move titanic 1.5 2.5 --speed=20
go run ./examples/naval --help
go run ./examples/props users add alice
```

## License

Copyright (c) 2025 Karagatan LLC. Licensed under BUSL-1.1.

## Contributions

If you find a bug or issue, please create a ticket. For now no external contributions are allowed.
