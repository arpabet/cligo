/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "go.arpabet.com/glue"

// Option configures badger using the functional options paradigm
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

// option that do nothing
func Nope() Option {
	return optionFunc(func(*implCliApplication) {
	})
}

// option that adds app name
func Name(name string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.name = name
	})
}

// option that adds app help
func Title(title string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.title = title
	})
}

// option that adds app help
func Help(help string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.help = help
	})
}

// option that adds version
func Version(version string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.version = version
	})
}

// option that adds build number
func Build(build string) Option {
	return optionFunc(func(a *implCliApplication) {
		a.build = build
	})
}

// option that adds verbose
func Verbose(verbose bool) Option {
	return optionFunc(func(a *implCliApplication) {
		a.verbose = verbose
	})
}

// option that adds beans
func Beans(beans ...interface{}) Option {
	return optionFunc(func(a *implCliApplication) {
		a.beans = append(a.beans, beans...)
	})
}

// option that adds properties in core context
func Properties(properties glue.Properties) Option {
	return optionFunc(func(a *implCliApplication) {
		a.properties = properties
	})
}
