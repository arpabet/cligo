/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"context"

	"go.arpabet.com/glue"
)

// Option configures a cligo application using the functional options paradigm
// popularized by Rob Pike and Dave Cheney. If you're unfamiliar with this style,
// see https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
type Option interface {
	apply(*implCliApplication)
}

// OptionFunc implements Option interface.
type optionFunc func(*implCliApplication)

// apply the configuration to the provided config.
func (fn optionFunc) apply(a *implCliApplication) {
	fn(a)
}

// Nope returns a no-op option (useful for conditional configuration).
func Nope() Option {
	return optionFunc(func(*implCliApplication) {
	})
}

// Name sets the application name (defaults to the binary name from os.Args[0]).
func Name(name string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.name = name
	})
}

// Title sets the display title shown in --version output.
func Title(title string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.title = title
	})
}

// Help sets the application description shown in help output.
func Help(help string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.help = help
	})
}

// Version sets the version string and enables the --version / -v flag.
func Version(version string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.version = version
	})
}

// Build sets the build identifier displayed alongside the version.
func Build(build string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.build = build
	})
}

// Verbose forces verbose mode on (also enabled by --verbose flag).
func Verbose(verbose bool) Option {
	return optionFunc(func(a *implCliApplication) {
		a.verbose = verbose
	})
}

// Beans registers groups, commands, and other DI beans with the application.
func Beans(beans ...interface{}) Option {
	return optionFunc(func(a *implCliApplication) {
		a.beans = append(a.beans, beans...)
	})
}

// Properties sets glue properties for dependency injection into the root container.
func Properties(properties glue.Properties) Option {
	return optionFunc(func(a *implCliApplication) {
		a.properties = properties
	})
}

// Context sets a custom base context for command execution.
// If not provided, Run() creates a signal-aware context that cancels on SIGINT/SIGTERM.
func Context(ctx context.Context) Option {
	return optionFunc(func(a *implCliApplication) {
		a.ctx = ctx
	})
}

// ConfigFile specifies a config file path to try loading into glue.Properties.
// Call multiple times to specify fallback paths — the first existing file is loaded.
// Supported formats (by extension): .properties, .yaml, .yml, .json, .toml.
// These are merged with any --config CLI flag values.
// Priority: flags > env vars > config file > defaults.
func ConfigFile(path string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.configFiles = append(a.configFiles, path)
	})
}

// Profile sets active glue profiles programmatically.
// These are merged with any --profile CLI flag values.
func Profile(profile string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.profiles = append(a.profiles, profile)
	})
}

// Color forces colored output on or off.
// By default, color is auto-detected: enabled for terminals, disabled for pipes.
// Respects the NO_COLOR environment variable (https://no-color.org/).
func Color(enabled bool) Option {
	return optionFunc(func(a *implCliApplication) {
		a.color = &enabled
	})
}
