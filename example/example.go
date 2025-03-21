package main

import (
	"fmt"
	"go.arpabet.com/cligo"
	"go.arpabet.com/glue"
)

type Cli struct {
	GroupField cligo.CliGroup `cli:"group=root"`
	Version    string         `cli:"app=version"`
}

func (g *Cli) Group() string {
	return "cli"
}

func (g *Cli) Help() string {
	return `Naval Fate.

	This is the docopt example adopted to Click but with some actual
	commands implemented and not just the empty parsing which really
	is not all that interesting.
	`
}

type Ship struct {
	GroupField cligo.CliGroup `cli:"group=cli"`
}

func (g *Ship) Group() string {
	return "ship"
}

func (g *Ship) Help() string {
	return `Manages ships.`
}

type ShipNew struct {
	GroupField cligo.CliGroup `cli:"group=ship"`
	Name       string         `cli:"argument=name"`
}

func (cmd *ShipNew) Command() string {
	return "new"
}

func (cmd *ShipNew) Help() string {
	return `Creates a new ship.`
}

func (cmd *ShipNew) Run(ctx cligo.Context) error {
	fmt.Printf("Created ship %s\n", cmd.Name)
	return nil
}

type ShipMove struct {
	GroupField cligo.CliGroup `cli:"group=ship"`
	Ship       string         `cli:"argument=ship"`
	X          float64        `cli:"argument=x"`
	Y          float64        `cli:"argument=y"`
	Speed      int            `cli:"option=speed,default=10,help=Speed in knots."`
	Verbose    bool           `cli:"option=verbose,default=false,help=Print verbose output."`
}

func (cmd *ShipMove) Command() string {
	return "move"
}

func (cmd *ShipMove) Help() string {
	return `Moves SHIP to the new location X,Y.`
}

func (cmd *ShipMove) Run(ctx cligo.Context) error {
	if cmd.Verbose {
		fmt.Printf("Moving ship %s to %v,%v with speed %d (verbose mode)\n", cmd.Ship, cmd.X, cmd.Y, cmd.Speed)
	} else {
		fmt.Printf("Moving ship %s to %v,%v with speed %d\n", cmd.Ship, cmd.X, cmd.Y, cmd.Speed)
	}
	return nil
}

func main() {

	context, err := glue.New(&Cli{},
		&Ship{},
		&ShipNew{},
		&ShipMove{})

	if err != nil {
		panic(err)
	}
	defer context.Close()

	cligo.Main(context)
}
