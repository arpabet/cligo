/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package main

import (
	"go.arpabet.com/cligo"
	"go.arpabet.com/glue"
)

type User struct {
	Parent cligo.CliGroup `cli:"group=cli"`
}

func (g *User) Group() string {
	return "users"
}

func (g *User) Help() (string, string) {
	return `Manages users.`, "Manages users by adding and removing them to/from the system."
}

type AddUser struct {
	Parent  cligo.CliGroup `cli:"group=users"`
	Name    string         `cli:"argument=name"`
	Profile string         `value:"profiles.active"`
}

func (cmd *AddUser) Command() string {
	return "add"
}

func (cmd *AddUser) Help() (string, string) {
	return "Adds user.", `This command adds user to the system.
It uses in this command to get the name of the user.`
}

func (cmd *AddUser) Run(ctx glue.Context) error {
	cligo.Echo("Add user '%s' in '%s' env", cmd.Name, cmd.Profile)
	return nil
}

type RemoveUser struct {
	Parent  cligo.CliGroup `cli:"group=users"`
	Name    string         `cli:"argument=name"`
	Profile string         `value:"profiles.active"`
}

func (cmd *RemoveUser) Command() string {
	return "remove"
}

func (cmd *RemoveUser) Help() (string, string) {
	return "Remove user.", `This command removes user from the system.
It uses in this command to get the name of the user.`
}

func (cmd *RemoveUser) Run(ctx glue.Context) error {
	cligo.Echo("Remove user '%s' in '%s' env", cmd.Name, cmd.Profile)
	return nil
}

func main() {

	banner := `
This is the basic CLI application example
based on simple commands related to user management.
`

	properties := glue.NewProperties()
	properties.Set("profiles.active", "dev")

	beans := []interface{}{
		&User{},
		&AddUser{},
		&RemoveUser{},
	}

	cligo.Main(cligo.Help(banner), cligo.Version("1.0.0"), cligo.Properties(properties), cligo.Beans(beans...))

}
