package cligo

import (
	"flag"
	"fmt"
	"go.arpabet.com/glue"
	"os"
	"reflect"
	"strings"
)

// Context holds the execution context for commands
type Context struct {
	Args []string
}

// Command represents a parsed CLI command
type Command struct {
	Name string
	Help string
	Run  func(ctx Context) error
}

// Group represents a command group
type Group struct {
	Name     string
	Help     string
	Commands map[string]Command
	Groups   map[string]*Group
}

// Runner interface for executable commands
type Runner interface {
	Run(ctx Context) error
}

// Registry to hold all registered groups and commands
type Registry struct {
	Root *Group
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		Root: &Group{
			Name:     "root",
			Commands: make(map[string]Command),
			Groups:   make(map[string]*Group),
		},
	}
}

// Register registers a struct as either a group or command
func (r *Registry) RegisterGroup(item interface{}) {
	t := reflect.TypeOf(item).Elem()

	// Get CLI annotations from the first field (assuming it's the group field)
	cliTag := t.Field(0).Tag.Get("cli")
	if cliTag == "" {
		return
	}

	annotations := parseAnnotations(cliTag)

	// Determine if it's a group or command
	parentGroup := strings.TrimPrefix(annotations["group"], "group=")

	// Find or create parent group
	currentGroup := r.Root
	if parentGroup != "root" && parentGroup != "" {
		if g, ok := r.Root.Groups[parentGroup]; ok {
			currentGroup = g
		} else {
			newGroup := &Group{
				Name:     parentGroup,
				Commands: make(map[string]Command),
				Groups:   make(map[string]*Group),
			}
			r.Root.Groups[parentGroup] = newGroup
			currentGroup = newGroup
		}
	}

	// Register as a group
	groupName := strings.TrimPrefix(annotations["group"], "group=")
	newGroup := &Group{
		Name:     groupName,
		Commands: make(map[string]Command),
		Groups:   make(map[string]*Group),
	}

	// Get help text from Help() method
	if helpMethod := reflect.ValueOf(item).MethodByName("Help"); helpMethod.IsValid() {
		newGroup.Help = helpMethod.Call(nil)[0].String()
	}

	currentGroup.Groups[groupName] = newGroup

}

func (r *Registry) RegisterCommand(item interface{}) {
	t := reflect.TypeOf(item).Elem()
	v := reflect.ValueOf(item).Elem() // Get the struct value

	// Get CLI annotations from the first field (assuming it's the group field)
	cliTag := t.Field(0).Tag.Get("cli")
	if cliTag == "" {
		return
	}

	annotations := parseAnnotations(cliTag)

	// Determine if it's a group or command
	parentGroup := strings.TrimPrefix(annotations["group"], "group=")

	// Find or create parent group
	currentGroup := r.Root
	if parentGroup != "root" && parentGroup != "" {
		if g, ok := r.Root.Groups[parentGroup]; ok {
			currentGroup = g
		} else {
			newGroup := &Group{
				Name:     parentGroup,
				Commands: make(map[string]Command),
				Groups:   make(map[string]*Group),
			}
			r.Root.Groups[parentGroup] = newGroup
			currentGroup = newGroup
		}
	}

	// Register as a command
	cmd := Command{
		Run: func(ctx Context) error {
			// Create a new instance and set arguments
			newInstance := reflect.New(t).Interface().(Runner)
			newValue := reflect.ValueOf(newInstance).Elem()

			// Set arguments from context
			for i := 1; i < t.NumField(); i++ {
				field := t.Field(i)
				if tag := field.Tag.Get("cli"); strings.HasPrefix(tag, "argument=") {
					if len(ctx.Args) > 0 {
						switch field.Type.Kind() {
						case reflect.String:
							newValue.Field(i).SetString(ctx.Args[0])
							ctx.Args = ctx.Args[1:]
						case reflect.Float64:
							var val float64
							fmt.Sscanf(ctx.Args[0], "%f", &val)
							newValue.Field(i).SetFloat(val)
							ctx.Args = ctx.Args[1:]
						}
					}
				}
			}

			return newInstance.Run(ctx)
		},
	}

	// Get command name and help
	if cmdMethod := reflect.ValueOf(item).MethodByName("Command"); cmdMethod.IsValid() {
		cmd.Name = cmdMethod.Call(nil)[0].String()
	}
	if helpMethod := reflect.ValueOf(item).MethodByName("Help"); helpMethod.IsValid() {
		cmd.Help = helpMethod.Call(nil)[0].String()
	}

	// Parse arguments and options
	for i := 1; i < t.NumField(); i++ {
		field := t.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue
		}

		anns := parseAnnotations(cliTag)
		switch {
		case strings.HasPrefix(cliTag, "argument="):
			// Arguments will be set in Run
		case strings.HasPrefix(cliTag, "option="):
			name := strings.TrimPrefix(anns["option"], "option=")
			defaultVal := anns["default"]
			help := anns["help"]

			switch field.Type.Kind() {
			case reflect.String:
				flag.StringVar(v.Field(i).Addr().Interface().(*string), name, defaultVal, help)
			case reflect.Int:
				intVal := 0
				if defaultVal != "" {
					fmt.Sscanf(defaultVal, "%d", &intVal)
				}
				flag.IntVar(v.Field(i).Addr().Interface().(*int), name, intVal, help)
			case reflect.Float64:
				floatVal := float64(0)
				if defaultVal != "" {
					fmt.Sscanf(defaultVal, "%f", &floatVal)
				}
				flag.Float64Var(v.Field(i).Addr().Interface().(*float64), name, floatVal, help)
			case reflect.Bool:
				boolVal := false
				if defaultVal == "true" {
					boolVal = true
				}
				flag.BoolVar(v.Field(i).Addr().Interface().(*bool), name, boolVal, help)
			}
		}
	}

	currentGroup.Commands[cmd.Name] = cmd

}

// Run parses arguments and executes the appropriate command
func (r *Registry) Run(ctx glue.Context) {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		r.printHelp(r.Root)
		return
	}

	current := r.Root
	var cmd Command
	var cmdArgs []string

	for i, arg := range args {
		if g, ok := current.Groups[arg]; ok {
			current = g
			continue
		}
		if c, ok := current.Commands[arg]; ok {
			cmd = c
			cmdArgs = args[i+1:]
			break
		}
	}

	if cmd.Run != nil {
		ctx := Context{Args: cmdArgs}
		if err := cmd.Run(ctx); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	r.printHelp(current)
}

// printHelp displays help information
func (r *Registry) printHelp(g *Group) {
	fmt.Printf("%s\n\n", g.Help)
	if len(g.Groups) > 0 {
		fmt.Println("Groups:")
		for name, group := range g.Groups {
			fmt.Printf("  %s: %s\n", name, group.Help)
		}
	}
	if len(g.Commands) > 0 {
		fmt.Println("Commands:")
		for name, cmd := range g.Commands {
			fmt.Printf("  %s: %s\n", name, cmd.Help)
		}
	}
}

// parseAnnotations parses CLI annotations from struct tags
func parseAnnotations(tag string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else {
			result[kv[0]] = ""
		}
	}
	return result
}

// Main entry point
func Main(ctx glue.Context) {
	registry := NewRegistry()

	// Register all groups
	for _, item := range ctx.Bean(CliGroupClass, 0) {
		registry.RegisterGroup(item.Object())
	}

	// Register all commands
	for _, item := range ctx.Bean(CliCommandClass, 0) {
		registry.RegisterCommand(item.Object())
	}

	registry.Run(ctx)
}
