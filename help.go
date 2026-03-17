/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "strings"

// printHelp prints help for a group
func (t *implCliApplication) printHelp(groupName string, stack []string) {

	groups := t.groups[groupName]
	commands := t.commands[groupName]

	path := strings.Join(stack, " ")

	if len(groups)+len(commands) > 0 {
		Echo("%s: %s %s [OPTIONS] COMMAND [ARGS]...", t.styled("Usage", ansiBold), t.name, path)
	} else {
		Echo("%s: %s %s [OPTIONS] [ARGS]...", t.styled("Usage", ansiBold), t.name, path)
	}

	help := t.helps[groupName]
	if help != "" {
		Echo("\n%s\n", help)
	}

	if groupName == RootGroup {
		Echo("%s:", t.styled("Options", ansiBold))
		if t.version != "" {
			Echo("  %s  Show the version and exit.", t.styled("-v, --version", ansiYellow))
		}
		Echo("  %s  Activate glue profiles (comma-separated).", t.styled("-p, --profile", ansiYellow))
		Echo("  %s   Load config file (repeatable).", t.styled("-c, --config", ansiYellow))
		Echo("  %s      Show extended logging information.", t.styled("--verbose", ansiYellow))
		Echo("  %s   Show this message and exit.", t.styled("-h, --help", ansiYellow))
		Echo("")
	}

	Echo("%s:", t.styled("Commands", ansiBold))
	for _, grp := range groups {
		if t.hidden[grp] {
			continue
		}
		shortDesc, _ := grp.Help()
		name := t.styled(grp.Group(), ansiCyan)
		if alias, ok := t.aliasOf[grp]; ok {
			name = name + " (" + alias + ")"
		}
		Echo("  %s\t%s", name, shortDesc)
	}

	for _, cmd := range commands {
		if t.hidden[cmd] {
			continue
		}
		shortDesc, _ := cmd.Help()
		name := t.styled(cmd.Command(), ansiCyan)
		if alias, ok := t.aliasOf[cmd]; ok {
			name = name + " (" + alias + ")"
		}
		Echo("  %s\t%s", name, shortDesc)
	}

}
