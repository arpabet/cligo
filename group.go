/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "fmt"

// RegisterGroup registers a command group
func (t *implCliApplication) RegisterGroup(group CliGroup) error {
	info := extractParentInfo(group)
	if info.group == "" {
		return fmt.Errorf("parent group not found in cli group: %v", group)
	}
	t.groups[info.group] = append(t.groups[info.group], group)
	if info.hidden {
		t.hidden[group] = true
	}
	if info.alias != "" {
		t.aliasOf[group] = info.alias
		if t.groupAliases[info.group] == nil {
			t.groupAliases[info.group] = make(map[string]CliGroup)
		}
		t.groupAliases[info.group][info.alias] = group
	}
	shortDesc, longDesc := group.Help()
	if len(longDesc) == 0 {
		longDesc = shortDesc
	}
	t.helps[group.Group()] = longDesc
	return nil
}

func (t *implCliApplication) findGroup(parentGroup, name string) CliGroup {
	for _, group := range t.groups[parentGroup] {
		if group.Group() == name {
			return group
		}
	}
	if aliasMap, ok := t.groupAliases[parentGroup]; ok {
		if group, ok := aliasMap[name]; ok {
			return group
		}
	}
	return nil
}
