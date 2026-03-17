/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"fmt"
	"reflect"
	"strings"
)

// getCommandUsage gets printable usage line
func (t *implCliApplication) getCommandUsage(cmd CliCommand, stack []string) string {

	// Print arguments and options
	cmdValue := reflect.ValueOf(cmd).Elem()
	cmdType := cmdValue.Type()

	// First get arguments
	var arguments []string
	for i := 0; i < cmdType.NumField(); i++ {
		field := cmdType.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue
		}

		tagParts := parseCliTag(cliTag)
		if argName, ok := tagParts["argument"]; ok {
			name := strings.ToUpper(argName)
			_, hasDefault := tagParts["default"]
			_, hasRequired := tagParts["required"]
			if hasDefault && !hasRequired {
				name = "[" + name + "]"
			}
			arguments = append(arguments, name)
		}
	}

	path := strings.Join(stack, " ")
	argsLine := strings.Join(arguments, " ")

	return fmt.Sprintf("%s: %s %s [OPTIONS] %s", t.styled("Usage", ansiBold), t.name, path, argsLine)
}

// getCommandTryUsage gets printable help with try statement
func (t *implCliApplication) getCommandTryUsage(cmd CliCommand, stack []string) string {
	path := strings.Join(stack, " ")
	return fmt.Sprintf("Try '%s %s --help' for help", t.name, path)
}
