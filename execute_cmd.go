/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"context"
	"fmt"
	"reflect"

	"github.com/spf13/pflag"
	"go.arpabet.com/glue"
)

// executeCommand parses arguments and options for a command and executes it
func (t *implCliApplication) executeCommand(ctx context.Context, c glue.Container, cmd CliCommand, args []string, stack []string) error {
	// Create a new value to store the parsed arguments
	cmdValue := reflect.ValueOf(cmd).Elem()
	cmdType := cmdValue.Type()

	// Prepare a custom flag set
	flagSet := pflag.NewFlagSet(cmd.Command(), pflag.ContinueOnError)
	flagSet.Usage = func() { t.printCommandHelp(cmd, stack) }

	// First pass: identify arguments and register options
	argDefs, options, envVars := t.identifyArgumentsAndOptions(cmdType, cmdValue, flagSet)

	// Add help option
	isHelp := flagSet.BoolP("help", "h", false, "Print help")
	isVerbose := flagSet.Bool("verbose", false, "Verbose output")

	// Parse flags
	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	argValues := flagSet.Args()

	if *isHelp {
		t.printCommandHelp(cmd, stack)
		return nil
	}

	// update verbose flag only if explicitly passed at command level
	if *isVerbose {
		t.verbose = true
	}

	// Set argument values
	err = t.setArgumentValues(argDefs, cmdValue, argValues, cmd, stack)
	if err != nil {
		return err
	}

	// Set option values: explicit flag > env var > default.
	t.setOptionValues(flagSet, options, envVars)

	cmdBeans, ok := t.commandBeans[cmd.Command()]
	if ok && len(cmdBeans) > 0 {
		child, err := c.Extend(cmdBeans...)
		if err != nil {
			Echo("%s\n%s\n", t.getCommandUsage(cmd, stack), t.getCommandTryUsage(cmd, stack))
			return fmt.Errorf("fail to initialize '%s' command scope context, %v", cmd.Command(), err)
		}
		defer child.Close()
		return cmd.Run(ctx, child)
	}

	// Execute the uknown command in the application context
	return cmd.Run(ctx, c)
}
