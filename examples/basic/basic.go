package main

import (
	"fmt"
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
	fmt.Printf("Created ship %s\n", cmd.Name)
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
		fmt.Printf("Moving ship %s to %v,%v with speed %d (verbose mode)\n", cmd.Ship, cmd.X, cmd.Y, cmd.Speed)
	} else {
		fmt.Printf("Moving ship %s to %v,%v with speed %d\n", cmd.Ship, cmd.X, cmd.Y, cmd.Speed)
	}
	return nil
}

func main() {

	banner := `Naval Fate.

	This is the docopt example adopted to cligo but with some actual
	commands implemented and not just the empty parsing which really
	is not all that interesting.
	`

	beans := []interface{}{
		&ShipNew{},
		&ShipMove{},
	}

	cligo.Main(cligo.Name("example"), cligo.Help(banner), cligo.Version("1.0.0"), cligo.Build("001"), cligo.Beans(beans))

}
