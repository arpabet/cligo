package cligo

import (
	"go.arpabet.com/glue"
	"reflect"
)

var CliGroupClass = reflect.TypeOf((*CliGroup)(nil)).Elem()

type CliGroup interface {
	// Group get group name
	Group() string
	// Help description about the group
	Help() (short string, optionalLong string)
}

var CliCommandClass = reflect.TypeOf((*CliCommand)(nil)).Elem()

type CliCommand interface {
	// Command get command name
	Command() string
	// Help description about the command
	Help() (short string, optionalLong string)
	// Run executes the command in context
	Run(ctx glue.Context) error
}

var CliCommandWithBeansClass = reflect.TypeOf((*CliCommandWithBeans)(nil)).Elem()

type CliCommandWithBeans interface {
	CliCommand
	// CommandBeans get optional beans for the command scope
	CommandBeans() []interface{}
}

var CliApplicationClass = reflect.TypeOf((*CliApplication)(nil)).Elem()

type CliApplication interface {
	Name() string
	Title() string
	Help() string
	Version() string
	Build() string
	Verbose() bool

	// Non-public method to keep beans private
	getBeans() []interface{}

	// RegisterGroup register the cli group in the context
	RegisterGroup(group CliGroup) error

	// RegisterCommand register the cli command in the context
	RegisterCommand(cmd CliCommand) error

	// RegisterCommandWithBeans register the cli command with beans in the context
	RegisterCommandWithBeans(cmd CliCommandWithBeans) error

	// Run CLI
	Execute(ctx glue.Context) error
}
