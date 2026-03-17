package cligo

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"go.arpabet.com/glue"
)

// implCliApplication is the main application structure
type implCliApplication struct {
	name         string
	title        string
	help         string
	version      string
	build        string
	verbose      bool
	color        *bool
	configFiles  []string
	profiles     []string
	ctx          context.Context
	beans        []interface{}
	properties   glue.Properties
	groups       map[string][]CliGroup
	commands     map[string][]CliCommand
	commandBeans map[string][]interface{}
	helps        map[string]string
	hidden       map[interface{}]bool
	aliasOf      map[interface{}]string
	cmdAliases   map[string]map[string]CliCommand
	groupAliases map[string]map[string]CliGroup
}

// New creates a new CLI application
func New(options ...Option) CliApplication {
	app := &implCliApplication{
		groups:       make(map[string][]CliGroup),
		commands:     make(map[string][]CliCommand),
		commandBeans: make(map[string][]interface{}),
		helps:        make(map[string]string),
		hidden:       make(map[interface{}]bool),
		aliasOf:      make(map[interface{}]string),
		cmdAliases:   make(map[string]map[string]CliCommand),
		groupAliases: make(map[string]map[string]CliGroup),
	}

	// first bean is application itself
	app.beans = []interface{}{app}

	// apply options
	for _, opt := range options {
		opt.apply(app)
	}

	if app.name == "" {
		app.name = filepath.Base(os.Args[0])
	}

	var str strings.Builder
	if app.title != "" {
		str.WriteString(app.title)
		str.WriteString("\n")
	}
	if app.help != "" {
		str.WriteString(app.help)
		str.WriteString("\n")
	}
	app.helps[RootGroup] = str.String()

	if !app.verbose {
		app.verbose = hasVerbose(os.Args[1:])
	}

	// Merge CLI --profile/-p flag values with programmatic profiles
	if cliProfiles := parseGlobalFlag(os.Args[1:], "profile", "p"); len(cliProfiles) > 0 {
		app.profiles = append(app.profiles, cliProfiles...)
	}

	// Merge CLI --config/-c flag values with programmatic config files
	if cliConfigs := parseGlobalFlag(os.Args[1:], "config", "c"); len(cliConfigs) > 0 {
		app.configFiles = append(app.configFiles, cliConfigs...)
	}

	return app
}

func (t *implCliApplication) Name() string {
	return t.name
}

func (t *implCliApplication) Title() string {
	return t.title
}

func (t *implCliApplication) Help() string {
	return t.help
}

func (t *implCliApplication) Version() string {
	return t.version
}

func (t *implCliApplication) Build() string {
	return t.build
}

func (t *implCliApplication) Verbose() bool {
	return t.verbose
}

func (t *implCliApplication) getBeans() []interface{} {
	return t.beans
}

func (t *implCliApplication) getProperties() glue.Properties {
	return t.properties
}

func (t *implCliApplication) getContext() context.Context {
	return t.ctx
}

func (t *implCliApplication) getConfigFiles() []string {
	return t.configFiles
}

func (t *implCliApplication) getProfiles() []string {
	return t.profiles
}
