/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.arpabet.com/glue"
)

// Execute parses arguments and runs the appropriate command
func (t *implCliApplication) Execute(ctx context.Context, c glue.Container) error {

	if len(os.Args) < 2 {
		t.printHelp(RootGroup, nil)
		return nil
	}

	// Check for version flag
	if t.version != "" {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			name := t.name
			if t.title != "" {
				name = t.title
			}
			if t.build != "" {
				Echo("%s Version %s Build %s", name, t.version, t.build)
			} else {
				Echo("%s Version %s", name, t.version)
			}
			if t.help != "" {
				Echo(t.help)
			}
			return nil
		}
	}

	// Check for help flag
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		t.printHelp(RootGroup, nil)
		return nil
	}

	var stack []string
	return t.parseAndExecute(ctx, c, RootGroup, os.Args[1:], stack)
}

// parseAndExecute recursively parses arguments and executes the appropriate command
func (t *implCliApplication) parseAndExecute(ctx context.Context, c glue.Container, currentGroup string, args []string, stack []string) error {
	if len(args) == 0 {
		t.printHelp(currentGroup, stack)
		return nil
	}

	// Check if the first argument is a group (by name or alias)
	matchedGroup := t.findGroup(currentGroup, args[0])
	if matchedGroup != nil {
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			t.printHelp(matchedGroup.Group(), stack)
			return nil
		}
		stack = append(stack, args[0])
		return t.parseAndExecute(ctx, c, matchedGroup.Group(), args[1:], stack)
	}

	// Check if the first argument is a command (by name or alias)
	matchedCmd := t.findCommand(currentGroup, args[0])
	if matchedCmd != nil {
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			t.printCommandHelp(matchedCmd, stack)
			return nil
		}
		stack = append(stack, args[0])
		return t.executeCommand(ctx, c, matchedCmd, args[1:], stack)
	}

	// Check if the first argument is a know option
	if args[0] == "--help" || args[0] == "-h" {
		t.printHelp(RootGroup, stack)
		return nil
	}

	if args[0] == "--verbose" {
		t.verbose = true
		t.printHelp(currentGroup, stack)
		return nil
	}

	if args[0] == "--profile" || args[0] == "-p" || strings.HasPrefix(args[0], "--profile=") || strings.HasPrefix(args[0], "-p=") ||
		args[0] == "--config" || args[0] == "-c" || strings.HasPrefix(args[0], "--config=") || strings.HasPrefix(args[0], "-c=") {
		// Skip --profile/--config (and short forms) and its value, then continue parsing
		skip := 1
		if (args[0] == "--profile" || args[0] == "-p" || args[0] == "--config" || args[0] == "-c") && len(args) > 1 {
			skip = 2
		}
		if skip >= len(args) {
			t.printHelp(currentGroup, stack)
			return nil
		}
		return t.parseAndExecute(ctx, c, currentGroup, args[skip:], stack)
	}

	t.printHelp(currentGroup, stack)
	if suggestion := t.suggest(currentGroup, args[0]); suggestion != "" {
		return fmt.Errorf("unknown command or group: %s. Did you mean %q?", args[0], suggestion)
	}
	return fmt.Errorf("unknown command or group: %s", args[0])
}
