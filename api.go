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
