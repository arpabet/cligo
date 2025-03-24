/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package main

import (
	"go.arpabet.com/cligo"
	"go.arpabet.com/glue"
)

type ShipNew struct {
	Parent cligo.CliGroup `cli:"group=cli"`
	Name   string         `cli:"argument=name"`
}

func (cmd *ShipNew) Command() string {
	return "new"
}

func (cmd *ShipNew) Help() (string, string) {
	return "Creates a new ship.", `This command creates a new ship.
It uses in order to place a ship in the game.`
}

func (cmd *ShipNew) Run(ctx glue.Context) error {
	cligo.Echo("Created ship %s", cmd.Name)
	return nil
}

type ShipMove struct {
	Parent  cligo.CliGroup `cli:"group=cli"`
	Ship    string         `cli:"argument=ship"`
	X       float64        `cli:"argument=x"`
	Y       float64        `cli:"argument=y"`
	Speed   int            `cli:"option=speed,default=10,help=Speed in knots."`
	Verbose bool           `cli:"option=verbose,default=false,help=Print verbose output."`
}

func (cmd *ShipMove) Command() string {
	return "move"
}

func (cmd *ShipMove) Help() (string, string) {
	return "Moves the ship", `Moves SHIP to the new location X,Y.`
}

func (cmd *ShipMove) Run(ctx glue.Context) error {
	if cmd.Verbose {
		cligo.Echo("Moving ship %s to %v,%v with speed %d (verbose mode)", cmd.Ship, cmd.X, cmd.Y, cmd.Speed)
	} else {
		cligo.Echo("Moving ship %s to %v,%v with speed %d", cmd.Ship, cmd.X, cmd.Y, cmd.Speed)
	}
	return nil
}

func main() {

	banner := `
This is the basic CLI application example
based on simple commands related to ship movement.
`

	beans := []interface{}{
		&ShipNew{},
		&ShipMove{},
	}

	cligo.Main(cligo.Name("basic"), cligo.Title("Basic CliGo Application"), cligo.Help(banner), cligo.Version("1.0.0"), cligo.Build("001"), cligo.Beans(beans...))

}
