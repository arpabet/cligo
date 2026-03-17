/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "fmt"

// RegisterCommand registers a command
func (t *implCliApplication) RegisterCommand(cmd CliCommand) error {
	info := extractParentInfo(cmd)
	if info.group == "" {
		return fmt.Errorf("parent group not found in cli command: %v", cmd)
	}
	t.commands[info.group] = append(t.commands[info.group], cmd)
	if info.hidden {
		t.hidden[cmd] = true
	}
	if info.alias != "" {
		t.aliasOf[cmd] = info.alias
		if t.cmdAliases[info.group] == nil {
			t.cmdAliases[info.group] = make(map[string]CliCommand)
		}
		t.cmdAliases[info.group][info.alias] = cmd
	}
	return nil
}

// RegisterCommandWithBeans registers a command with beans
func (t *implCliApplication) RegisterCommandWithBeans(cmd CliCommandWithBeans) error {
	info := extractParentInfo(cmd)
	if info.group == "" {
		return fmt.Errorf("parent group not found in cli command: %v", cmd)
	}
	t.commands[info.group] = append(t.commands[info.group], cmd)
	if info.hidden {
		t.hidden[cmd] = true
	}
	if info.alias != "" {
		t.aliasOf[cmd] = info.alias
		if t.cmdAliases[info.group] == nil {
			t.cmdAliases[info.group] = make(map[string]CliCommand)
		}
		t.cmdAliases[info.group][info.alias] = cmd
	}

	commandBeans := cmd.CommandBeans()
	if len(commandBeans) > 0 {
		t.commandBeans[cmd.Command()] = append(t.commandBeans[cmd.Command()], commandBeans...)
	}
	return nil
}

func (t *implCliApplication) findCommand(parentGroup, name string) CliCommand {
	for _, cmd := range t.commands[parentGroup] {
		if cmd.Command() == name {
			return cmd
		}
	}
	if aliasMap, ok := t.cmdAliases[parentGroup]; ok {
		if cmd, ok := aliasMap[name]; ok {
			return cmd
		}
	}
	return nil
}
