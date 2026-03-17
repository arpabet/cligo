/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"fmt"
	"reflect"
	"strings"
)

// printCommandHelp prints help for a specific command
func (t *implCliApplication) printCommandHelp(cmd CliCommand, stack []string) {

	// Print arguments and options
	cmdValue := reflect.ValueOf(cmd).Elem()
	cmdType := cmdValue.Type()

	Echo(t.getCommandUsage(cmd, stack))

	shortDesc, longDesc := cmd.Help()
	if len(longDesc) == 0 {
		longDesc = shortDesc
	}

	Echo("%s\n", longDesc)

	// Print argument details
	t.printArgumentDetails(cmdType)

	// Finally print option details
	t.printOptionDetails(cmdType)
}

func (t *implCliApplication) printArgumentDetails(cmdType reflect.Type) {
	var argLines []string
	for i := 0; i < cmdType.NumField(); i++ {
		field := cmdType.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue
		}

		tagParts := parseCliTag(cliTag)
		if argName, ok := tagParts["argument"]; ok {
			help := tagParts["help"]
			if help == "" {
				help = fmt.Sprintf("%s argument", argName)
			}
			_, hasDefault := tagParts["default"]
			_, hasRequired := tagParts["required"]
			if hasDefault && !hasRequired {
				help = help + fmt.Sprintf(" [default: %s]", tagParts["default"])
			} else {
				help = help + " [required]"
			}
			argLines = append(argLines, fmt.Sprintf("  %s\t%s", t.styled(strings.ToUpper(argName), ansiGreen), help))
		}
	}
	if len(argLines) > 0 {
		Echo("%s:", t.styled("Arguments", ansiBold))
		for _, line := range argLines {
			fmt.Println(line)
		}
		fmt.Println()
	}
}

func (t *implCliApplication) printOptionDetails(cmdType reflect.Type) {
	var hasOptions bool
	for i := 0; i < cmdType.NumField(); i++ {
		field := cmdType.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue
		}

		tagParts := parseCliTag(cliTag)
		if optName, ok := tagParts["option"]; ok {
			if !hasOptions {
				Echo("%s:", t.styled("Options", ansiBold))
				hasOptions = true
			}

			defaultVal := tagParts["default"]
			help := tagParts["help"]
			if help == "" {
				help = fmt.Sprintf("%s option", optName)
			}

			defaultText := ""
			if defaultVal != "" {
				defaultText = fmt.Sprintf(" [default: %s]", defaultVal)
			}

			envText := ""
			if envVar, ok := tagParts["env"]; ok {
				envText = fmt.Sprintf(" [$%s]", envVar)
			}

			fmt.Printf("  %s  %s%s%s\n", t.styled("--"+optName, ansiYellow), help, defaultText, envText)
		}
	}
}
