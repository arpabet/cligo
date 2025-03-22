package cligo

// Option configures badger using the functional options paradigm
// popularized by Rob Pike and Dave Cheney. If you're unfamiliar with this style,
// see https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
type Option interface {
	apply(*App)
}

// OptionFunc implements Option interface.
type optionFunc func(*App)

// apply the configuration to the provided config.
func (fn optionFunc) apply(a *App) {
	fn(a)
}

// option that do nothing
func Nope() Option {
	return optionFunc(func(*App) {
	})
}

// option that adds app name
func Name(name string) Option {
	return optionFunc(func(a *App) {
		a.name = name
	})
}

// option that adds app help
func Title(title string) Option {
	return optionFunc(func(a *App) {
		a.title = title
	})
}

// option that adds app help
func Help(help string) Option {
	return optionFunc(func(a *App) {
		a.help = help
	})
}

// option that adds version
func Version(version string) Option {
	return optionFunc(func(a *App) {
		a.version = version
	})
}

// option that adds build number
func Build(build string) Option {
	return optionFunc(func(a *App) {
		a.build = build
	})
}

// option that adds verbose
func Verbose(verbose bool) Option {
	return optionFunc(func(a *App) {
		a.verbose = verbose
	})
}

// option that adds beans
func Beans(beans ...interface{}) Option {
	return optionFunc(func(a *App) {
		a.beans = append(a.beans, beans...)
	})
}
