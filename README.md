# cligo
CLI Library on Golang

Simple CLI Library for the golang.

### Usage

Define the CLI group.
The root groups should use 'group=cli' as reference to root ones.
The subgroups should use 'group=<parent_group>'.
```
type Ship struct {
	Parent cligo.CliGroup `cli:"group=cli"`
}

func (g *Ship) Group() string {
	return "ship"
}

func (g *Ship) Help() (string, string) {
	return `Manages ships.`, ""
}
```

Define the command
```
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
```

After that create the main class
```
func main() {

	beans := []interface{}{
		&Ship{},
		&ShipNew{},
	}

	help := `Naval Fate.`

	cligo.Main(cligo.Beans(beans), cligo.Help(help))

}
```

That's it, very similar to python click library.

Issues and contributions are welcome.