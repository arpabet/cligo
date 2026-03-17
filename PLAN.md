# Cligo Improvement Plan

## Phase 1: Fix Bugs & Foundation (Critical)

- [x] **1.1** Fix `hasArgs` dead code in `printCommandHelp` — scan fields to actually set `hasArgs = true`
- [x] **1.2** Fix `-v` conflict — `--verbose` has no short flag; `-V` is the short flag for `--version`
- [x] **1.3** Wire short flags — use pflag's `StringP`/`IntP`/`BoolP` when `short` tag is present
- [x] **1.4** Fix verbose overwrite — only update `app.verbose` if `--verbose` was explicitly passed
- [x] **1.5** Fix README — update `glue.Context` → `glue.Container`
- [ ] **1.6** Add tests — unit tests for tag parsing, argument parsing, command execution, help generation, error cases. Target 80%+ coverage.
- [ ] **1.7** Upgrade Go minimum — move to at least `go 1.21` for `slog`, better generics
- [x] **1.8** Replace `pkg/errors` — use `fmt.Errorf` with `%w` (stdlib since Go 1.13)

## Phase 2: Essential Features

- [x] **2.1** `context.Context` support — add `context.Context` to command execution path, wire to DI container. Enables timeouts, cancellation, and signal handling.
- [x] **2.2** Signal handling — trap SIGINT/SIGTERM, propagate via context cancellation.
- [x] **2.3** Required vs optional arguments — add `cli:"argument=name,required"` tag support; allow optional args with defaults.
- [x] **2.4** Environment variable binding — `cli:"option=port,env=APP_PORT"` reads from env if flag not provided.
- [ ] **2.5** Validation interface — `Validate() error` method on commands, called after parsing before `Run()`.
- [ ] **2.6** Middleware/hooks — `BeforeRun(c glue.Container) error` and `AfterRun(c glue.Container) error` interfaces. Also a global middleware chain on `CliApplication`.
- [x] **2.7** Slice arguments — support `[]string`, `[]int` for options that can be repeated (`--tag=foo --tag=bar`).

## Phase 3: Polish & Ecosystem

- [ ] **3.1** Shell completion generation — generate bash/zsh/fish completion scripts from the command tree.
- [x] **3.2** Hidden commands — `cli:"hidden"` tag to exclude from help but still executable.
- [x] **3.3** Command aliases — `cli:"alias=mv"` for alternate names.
- [x] **3.4** Colored/formatted output — optional ANSI color support for help and errors.
- [ ] **3.5** Man page generation — generate man pages from command definitions.
- [ ] **3.6** Testing helpers — `cligo.TestRun(args []string, beans ...interface{}) error` for integration testing without `os.Args`.
- [x] **3.7** Config file support — integrate with glue properties to load from YAML/TOML/env files automatically.
- [x] **3.8** Profile support — `--profile` flag and `Profiles()` option to activate glue profiles for environment-aware bean registration.
- [ ] **3.9** Plugin system — allow loading commands from external binaries or Go plugins.

## Phase 4: Developer Experience

- [ ] **4.1** Better error messages — show "did you mean X?" for typos using Levenshtein distance.
- [ ] **4.2** Command groups in help — allow grouping commands under headers in help output (e.g., "Database Commands:", "Auth Commands:").
- [ ] **4.3** Auto-generated usage examples — `Examples() string` method on commands.
- [ ] **4.4** Deprecation support — `cli:"deprecated=Use X instead"` tag.
- [ ] **4.5** Structured logging — replace `log.Default()` with `slog` for structured verbose output.
- [ ] **4.6** CI pipeline — GitHub Actions for test, lint, vet on PRs.
