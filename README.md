# cligo

**The CLI framework that thinks like your application, not just your terminal.**

Cligo is a declarative CLI framework for Go that brings the power of dependency injection, struct-based declarations, and enterprise configuration patterns to command-line applications. Inspired by Python's [Click](https://click.palletsprojects.com/) and built on top of the [glue](https://go.arpabet.com/glue) DI framework, cligo eliminates the boilerplate of wiring commands, parsing config files, and managing environments -- so you can focus on what your commands actually do.

## Why cligo?

Most Go CLI frameworks stop at flag parsing. Cligo goes further:

- **Declare, don't wire.** Define commands as Go structs with `cli` tags. No manual flag registration, no boilerplate factories.
- **Dependency injection built in.** Every command receives a DI container. Services, database connections, and configuration are injected automatically -- not threaded through closures or globals.
- **Spring-like profiles.** Activate `dev`, `staging`, or `prod` configurations with `--profile`. Swap entire service implementations per environment with a single flag -- no other Go CLI framework offers this.
- **Config files out of the box.** Load `.properties`, `.yaml`, `.json`, `.toml`, or `.env` files without a separate library. Values flow through DI and are available everywhere.
- **One binary, zero glue code.** No Viper, no separate config library, no manual binding. Cligo unifies CLI parsing, configuration, DI, and execution into a single coherent framework.

## How cligo compares

| Feature | cligo | Cobra | urfave/cli | Kong | go-arg |
|---------|:-----:|:-----:|:----------:|:----:|:------:|
| **Struct-tag declarations** | Yes | -- | -- | Yes | Yes |
| **Dependency injection** | Yes (glue) | -- | -- | Partial (Bind) | -- |
| **Config files (built-in)** | Yes (5 formats) | Via Viper | Via altsrc | JSON only | -- |
| **Profile system** | Yes | -- | -- | -- | -- |
| **.env file support** | Yes | Via Viper | Via altsrc | -- | -- |
| **Subcommands / groups** | Yes | Yes | Yes | Yes | Yes |
| **Command aliases** | Yes | Yes | Yes | Yes | -- |
| **Hidden commands** | Yes | Yes | Yes | Yes | -- |
| **Env var binding** | Yes | Via Viper | Yes | Yes | Yes |
| **Slice / repeated options** | Yes | Yes | Yes | Yes | Yes |
| **Colored help output** | Yes (auto) | -- | -- | -- | -- |
| **Panic recovery** | Yes | -- | -- | -- | -- |
| **context.Context** | Yes (signal-aware) | Yes | Yes | Own context | -- |
| **"Did you mean?" typo hints** | Yes | Yes | Partial | -- | -- |
| **Command-scoped beans** | Yes | -- | -- | -- | -- |
| **Optional/required args with defaults** | Yes | Partial | Yes | Yes | Yes |
| **Short flags (-v, -h)** | Yes | Yes | Yes | Yes | Yes |
| **Functional options API** | Yes | -- | -- | Yes | -- |
| **Signal-aware graceful shutdown** | Yes (built-in) | -- | -- | -- | -- |

**Key advantages over alternatives:**

- **vs Cobra:** Cobra requires Viper for config files, env binding, and .env support -- three separate libraries to assemble. Cligo provides all of this in one import. Cobra also lacks struct-tag-based declarations, DI, profiles, colored help, and panic recovery.
- **vs urfave/cli:** Requires `cli-altsrc` for config files. No DI, no profiles, no struct tags, no colored help. Flag definitions are verbose typed structs rather than declarative tags.
- **vs Kong:** The closest alternative for struct-based parsing and DI. However, Kong's DI is limited to `Bind()`/`BindTo()` -- it is not a full container with lifecycle management. Kong lacks built-in config file loading (beyond JSON), profiles, colored help, panic recovery, and signal handling.
- **vs go-arg:** Clean struct-tag parsing but no DI, no config files, no subcommand aliases, no hidden commands, no colored help, no error suggestions.

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

### Hidden Commands

Add `hidden` to the parent field tag to exclude a command or group from help output while keeping it executable:

```go
type Debug struct {
    Parent cligo.CliGroup `cli:"group=cli,hidden"`
}
```

### Aliases

Add `alias=<name>` to the parent field tag to define an alternate name for a command or group:

```go
type ShipMove struct {
    Parent cligo.CliGroup `cli:"group=ship,alias=mv"`
}
```

Aliases work for both commands and groups. The alias is shown in help output as `move (mv)`.

### Arguments

Positional arguments are declared with `cli:"argument=<name>"` struct tags. They are parsed in the order they appear in the struct. Arguments are required by default; add `default=<value>` to make them optional:

```go
type Move struct {
    Parent cligo.CliGroup `cli:"group=ship"`
    Ship   string         `cli:"argument=ship"`           // required (default)
    X      float64        `cli:"argument=x,required"`     // explicitly required
    Y      float64        `cli:"argument=y,default=0.0"`  // optional with default
}
```

```
$ app ship move titanic 1.5       # Y defaults to 0.0
$ app ship move titanic 1.5 2.5   # Y explicitly set
```

Supported argument types: `string`, `int` (all sizes), `float32`, `float64`.

### Options

Named flags are declared with `cli:"option=<name>"` and support defaults, help text, short flags, and environment variable binding:

```go
type Move struct {
    Parent cligo.CliGroup `cli:"group=ship"`
    Ship   string         `cli:"argument=ship"`
    Speed  int            `cli:"option=speed,short=-s,default=10,help=Speed in knots"`
    Dry    bool           `cli:"option=dry,default=false,help=Dry run mode"`
    Label  string         `cli:"option=label,default=unnamed,help=Ship label"`
    Port   int            `cli:"option=port,default=8080,env=APP_PORT,help=Port number"`
}
```

```
$ app ship move titanic --speed=20
$ app ship move titanic -s 20
$ app ship move titanic --dry --label=flagship
$ APP_PORT=3000 app ship move titanic   # port from env
```

Option value priority: explicit flag > environment variable > default value.

Supported option types: `string`, `int` (all sizes), `float32`, `float64`, `bool`.

### Slice Options

Options with slice types (`[]string`, `[]int`, `[]float64`, `[]bool`) can be repeated to collect multiple values:

```go
type Build struct {
    Parent cligo.CliGroup `cli:"group=cli"`
    Tags   []string       `cli:"option=tag,short=-t,env=BUILD_TAGS,help=Add a tag"`
    Ports  []int          `cli:"option=port,help=Expose ports"`
}
```

```
$ app build --tag=v1 --tag=latest --port=8080 --port=9090
$ app build -t v1 -t latest
$ BUILD_TAGS=v1,latest app build   # env var: comma-separated
```

For environment variables, slice values are comma-separated. CLI flags always take priority over environment variables.

## Struct Tag Reference

All metadata is declared in the `cli` struct tag with comma-separated `key=value` pairs:

| Tag | Description | Example |
|-----|-------------|---------|
| `group=<name>` | Parent group (required on the `CliGroup` field) | `cli:"group=cli"` |
| `argument=<name>` | Positional argument (required by default) | `cli:"argument=name"` |
| `option=<name>` | Named flag/option | `cli:"option=speed"` |
| `required` | Explicitly marks an argument as required | `cli:"argument=name,required"` |
| `short=-<char>` | Single-character short flag | `cli:"option=speed,short=-s"` |
| `default=<value>` | Default value for an option or argument | `cli:"argument=y,default=0.0"` |
| `help=<text>` | Help text for an option | `cli:"option=speed,help=Speed in knots"` |
| `env=<VAR>` | Environment variable fallback for an option | `cli:"option=port,env=APP_PORT"` |
| `hidden` | Hide command/group from help output (still executable) | `cli:"group=cli,hidden"` |
| `alias=<name>` | Alternate name for a command or group | `cli:"group=ship,alias=mv"` |

Tags can be combined: `cli:"option=speed,short=-s,default=10,env=SPEED,help=Speed in knots"`

Supported types for arguments: `string`, `int` (all sizes), `float32`, `float64`.
Supported types for options: `string`, `int` (all sizes), `float32`, `float64`, `bool`, `[]string`, `[]int`, `[]float64`, `[]bool`.

## Application Options

Configure the application using functional options passed to `Main()` or `Run()`:

```go
cligo.Main(
    cligo.Name("myapp"),          // Binary name (defaults to os.Args[0])
    cligo.Title("My Application"),// Display title shown in --version
    cligo.Help("Description."),   // Help text shown in usage
    cligo.Version("1.0.0"),       // Enables --version / -v flag
    cligo.Build("abc123"),        // Build identifier shown alongside version
    cligo.Verbose(true),          // Force verbose mode
    cligo.Context(ctx),           // Custom context (defaults to signal-aware context)
    cligo.Color(true),            // Force colored output (auto-detected by default)
    cligo.Beans(beans...),        // Register groups and commands
    cligo.Properties(props),      // Glue properties for DI
)
```

| Option | Description |
|--------|-------------|
| `Name(s)` | Application name (defaults to binary name) |
| `Title(s)` | Display title for version output |
| `Help(s)` | Description shown in help output |
| `Version(s)` | Version string; enables `--version` / `-v` |
| `Build(s)` | Build identifier shown with version |
| `Verbose(b)` | Force verbose mode on |
| `Context(ctx)` | Custom `context.Context` (defaults to signal-aware context) |
| `Color(b)` | Force colored output on/off (auto-detected by default, respects `NO_COLOR`) |
| `ConfigFile(path)` | Load config file (repeatable, merged with `--config` flag) |
| `Profile(p)` | Activate glue profile (repeatable, merged with `--profile` flag) |
| `Beans(b...)` | Groups, commands, and other DI beans |
| `Properties(p)` | Glue properties for dependency injection |
| `Nope()` | No-op (useful for conditional options) |

## Global Flags

These flags are handled automatically:

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help for the application, group, or command |
| `--version`, `-v` | Show version and build info (requires `Version()` option) |
| `--profile`, `-p` | Activate glue profiles (comma-separated, repeatable) |
| `--config`, `-c` | Load config file (repeatable, merged with `ConfigFile()` option) |
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

### Profiles

Cligo integrates with glue's profile system for environment-aware bean registration and configuration. Profiles can be activated via the `--profile` CLI flag, the `Profile()` option, or the `glue.profiles.active` property.

```bash
# Single profile
myapp --profile dev serve

# Multiple profiles (comma-separated or repeated)
myapp --profile dev,local serve
myapp --profile dev --profile local serve

# Short flag
myapp -p staging serve

# Equals form
myapp --profile=staging serve
```

Programmatic activation with `Profile()` is merged with CLI flag values:

```go
cligo.Main(
    cligo.Profile("base"),
    cligo.Beans(
        glue.IfProfile("dev", &devDB{}),
        glue.IfProfile("prod", &prodDB{}),
        glue.IfProfile("dev|staging", &debugEndpoint{}),
        glue.IfProfile("!prod", &mockMetrics{}),
        &ServeCmd{},
    ),
)
```

Profile expressions supported by glue:

| Expression | Meaning |
|------------|---------|
| `"dev"` | Active when `dev` profile is active |
| `"!prod"` | Active when `prod` is NOT active |
| `"dev\|staging"` | Active when either `dev` OR `staging` is active |
| `"dev&local"` | Active when both `dev` AND `local` are active |

### Config Files

Use `ConfigFile()` to specify fallback config file paths -- the first existing file is loaded via `glue.PropertySource`. Format is detected by extension. Config files can also be specified from the command line with `--config`:

```go
cligo.Main(
    cligo.ConfigFile("config.properties"),
    cligo.ConfigFile("config.yaml"),
    cligo.ConfigFile("config.json"),
    cligo.Beans(&AddUser{}),
)
```

```bash
# Override config from CLI (repeatable, merged with ConfigFile options)
myapp --config /etc/myapp/config.yaml serve
myapp --config base.properties --config override.properties serve
```

Supported formats:

| Extension | Format | Example |
|-----------|--------|---------|
| `.properties` | Java properties (native glue format) | `app.profile = dev` |
| `.yaml`, `.yml` | YAML (nested keys flattened with dots) | `app:\n  profile: dev` |
| `.json` | JSON (nested keys flattened with dots) | `{"app": {"profile": "dev"}}` |
| `.toml` | TOML | `[app]\nprofile = "dev"` |

For `.env` files, register `glue.DotEnvPropertyResolver{}` as a bean:

```go
cligo.Main(
    cligo.ConfigFile(".env"),
    cligo.Beans(&glue.DotEnvPropertyResolver{}, &AddUser{}),
)
```

YAML and JSON nested structures are flattened with dot notation:

```yaml
# config.yaml → app.db.host=localhost, app.db.port=5432
app:
  db:
    host: localhost
    port: 5432
```

Config values are merged into `glue.Properties` before the DI container is created, so they're available via `value:"key"` struct tags. Priority: flags > env vars > config file > defaults.

## Error Handling

Cligo provides robust error handling out of the box:

- **Panic recovery:** Panics during command execution are caught and converted to errors, keeping the application stable.
- **Typo suggestions:** Unknown commands trigger "Did you mean X?" hints using Levenshtein distance matching, covering commands, groups, and aliases.
- **Helpful error messages:** Missing arguments, invalid types, and unknown commands all produce messages with usage hints and `--help` pointers.

```
$ app shp new titanic
Error: unknown command or group: shp. Did you mean "ship"?
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
