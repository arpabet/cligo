# cligo examples

Examples demonstrating the [cligo](https://go.arpabet.com/cligo) CLI framework.

## basic

Flat commands registered directly under the root group. Demonstrates arguments, typed options, version/build info, and the `Echo` helper.

```
go run ./examples/basic new titanic
go run ./examples/basic move titanic 1.5 2.5 --speed=20
go run ./examples/basic --version
go run ./examples/basic --help
```

## naval

Nested command groups (`ship`, `mine`) with multiple commands under each. Demonstrates the full group hierarchy, short flags (`-s`), and boolean options.

```
go run ./examples/naval ship new enterprise
go run ./examples/naval ship move enterprise 3.0 4.0 -s 25
go run ./examples/naval ship shoot enterprise 5.0 6.0
go run ./examples/naval mine set 10.0 20.0 --drifting
go run ./examples/naval mine remove 10.0 20.0
go run ./examples/naval --help
go run ./examples/naval ship --help
```

## props

Property injection via `glue.Properties`. Demonstrates injecting configuration values (like an active profile) into command struct fields using glue's `value` tag.

```
go run ./examples/props users add alice
go run ./examples/props users remove bob
go run ./examples/props --help
```
