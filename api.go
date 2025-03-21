package cligo

import "reflect"

var CliGroupClass = reflect.TypeOf((*CliGroup)(nil)).Elem()

type CliGroup interface {
	// Group get group name
	Group() string
	// Help description about the group
	Help() string
}

var CliCommandClass = reflect.TypeOf((*CliCommand)(nil)).Elem()

type CliCommand interface {
	// Command get command name
	Command() string
	// Help description about the command
	Help() string
	// Run executes the command in context
	Run(ctx Context) error
}
