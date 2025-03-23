/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package main

import (
	"go.arpabet.com/cligo"
	"go.arpabet.com/glue"
)

type Ship struct {
	Parent cligo.CliGroup `cli:"group=cli"`
}

func (g *Ship) Group() string {
	return "ship"
}

func (g *Ship) Help() (string, string) {
	return `Manages ships.`, ""
}

type ShipNew struct {
	Parent cligo.CliGroup `cli:"group=ship"`
	Name   string         `cli:"argument=name"`
}

func (cmd *ShipNew) Command() string {
	return "new"
}

func (cmd *ShipNew) Help() (string, string) {
	return `Creates a new ship.`, ""
}

func (cmd *ShipNew) Run(ctx glue.Context) error {
	cligo.Echo("Created ship %s", cmd.Name)
	return nil
}

type ShipMove struct {
	Parent  cligo.CliGroup `cli:"group=ship"`
	Ship    string         `cli:"argument=ship"`
	X       float64        `cli:"argument=x"`
	Y       float64        `cli:"argument=y"`
	Speed   int            `cli:"option=speed,short=-s,default=10,help=Speed in knots."`
	Verbose bool           `cli:"option=verbose,default=false,help=Print verbose output."`
}

func (cmd *ShipMove) Command() string {
	return "move"
}

func (cmd *ShipMove) Help() (string, string) {
	return `Moves SHIP to the new location X,Y.`, ""
}

func (cmd *ShipMove) Run(ctx glue.Context) error {
	if cmd.Verbose {
		cligo.Echo("Moving ship %s to %v,%v with speed %d (verbose mode)", cmd.Ship, cmd.X, cmd.Y, cmd.Speed)
	} else {
		cligo.Echo("Moving ship %s to %v,%v with speed %d", cmd.Ship, cmd.X, cmd.Y, cmd.Speed)
	}
	return nil
}

type Shoot struct {
	Parent cligo.CliGroup `cli:"group=ship"`
	Ship   string         `cli:"argument=ship"`
	X      float64        `cli:"argument=x"`
	Y      float64        `cli:"argument=y"`
}

func (cmd *Shoot) Command() string {
	return "shoot"
}

func (cmd *Shoot) Help() (string, string) {
	return `Makes SHIP fire to X,Y.`, ""
}

func (cmd *Shoot) Run(ctx glue.Context) error {
	cligo.Echo("Ship %s fires to %v,%v", cmd.Ship, cmd.X, cmd.Y)
	return nil
}

type Mine struct {
	Parent cligo.CliGroup `cli:"group=cli"`
}

func (g *Mine) Group() string {
	return "mine"
}

func (g *Mine) Help() (string, string) {
	return `Manages mines.`, ""
}

type Set struct {
	Parent   cligo.CliGroup `cli:"group=mine"`
	X        float64        `cli:"argument=x"`
	Y        float64        `cli:"argument=y"`
	Moored   bool           `cli:"option=moored,default=true,help=Moored (anchored) mine. Default."`
	Drifting bool           `cli:"option=drifting,help=Drifting mine.."`
}

func (cmd *Set) Command() string {
	return "set"
}

func (cmd *Set) Help() (string, string) {
	return `Makes SHIP fire to X,Y.`, ""
}

func (cmd *Set) Run(ctx glue.Context) error {
	cmd.Moored = !cmd.Drifting
	cligo.Echo("Set %v mine at %v,%v", cmd.Moored, cmd.X, cmd.Y)
	return nil
}

type Remove struct {
	Parent cligo.CliGroup `cli:"group=mine"`
	X      float64        `cli:"argument=x"`
	Y      float64        `cli:"argument=y"`
}

func (cmd *Remove) Command() string {
	return "remove"
}

func (cmd *Remove) Help() (string, string) {
	return `Removes a mine at a specific coordinate.`, ""
}

func (cmd *Remove) Run(ctx glue.Context) error {
	cligo.Echo("Removed mine at %v,%v", cmd.X, cmd.Y)
	return nil
}

func main() {

	beans := []interface{}{
		&Ship{},
		&ShipNew{},
		&ShipMove{},
		&Shoot{},
		&Mine{},
		&Set{},
		&Remove{},
	}

	help := `Naval Fate.
		This is the docopt example adopted to cligo but with some actual
		commands implemented and not just the empty parsing which really
		is not all that interesting.
	`

	cligo.Main(cligo.Beans(beans), cligo.Help(help))

}
