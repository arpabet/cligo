package main

import (
	"fmt"
	"go.arpabet.com/cligo"
	"go.arpabet.com/glue"
)

type CliGroup interface {
	// Group get group name
	Group() string
	// Help description about the group
	Help() string
}

type CliCommand interface {
	// Command get command name
	Command() string
	// Help description about the command
	Help() string
	// Run executes the command in context
	Run(ctx glue.Context) error
}

type Cli struct {
	group   CliGroup `cli:"group=root"`
	version string   `cli:"app=version"`
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
	group CliGroup `cli:"group=cli"`
}

func (g *Ship) Group() string {
	return "ship"
}

func (g *Ship) Help() string {
	return `Manages ships.`
}

type ShipNew struct {
	group   CliGroup `cli:"group=ship"`
	name    string   `cli:"argument=name"`
	count   int      `cli:"option=count,default=1,help=number of greetings"`
	verbose bool     `cli:"option=verbose,default=false,help=Print verbose output."`
}

func (cmd *ShipNew) Command() string {
	return "new"
}

func (cmd *ShipNew) Help() string {
	return `Creates a new ship.`
}

func (cmd *ShipNew) Run(ctx glue.Context) error {
	fmt.Printf("Created ship %s", cmd.name)
	return nil
}

type ShipMove struct {
	group   CliGroup `cli:"group=ship"`
	ship    string   `cli:"argument=ship"`
	x       float64  `cli:"argument=x"`
	y       float64  `cli:"argument=x"`
	speed   int      `cli:"option=speed,default=10,help=Speed in knots."`
	verbose bool     `cli:"option=verbose,default=false,help=Print verbose output."`
}

func (cmd *ShipMove) Command() string {
	return "move"
}

func (cmd *ShipMove) Help() string {
	return `Moves SHIP to the new location X,Y.`
}

func (cmd *ShipMove) Run(ctx glue.Context) error {
	if cmd.verbose {
		fmt.Printf("Moving ship %s to %v,%v with speed %d (verbose mode)\n", cmd.ship, cmd.x, cmd.y, cmd.speed)
	} else {
		fmt.Printf("Moving ship %s to %v,%v with speed %d\n", cmd.ship, cmd.x, cmd.y, cmd.speed)
	}
	return nil
}

func main() {
	cligo.Main(
		&Cli{},
		&Ship{},
		&ShipNew{},
		&ShipMove{},
	)
}
