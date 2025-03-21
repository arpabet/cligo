package cligo

import (
	"flag"
	"fmt"
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
	Name    string
	Help    string
	Execute func(ctx Context) error
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
func (r *Registry) Register(item interface{}) {
	t := reflect.TypeOf(item).Elem()
	v := reflect.ValueOf(item)

	// Get CLI annotations
	cliTag := t.Field(0).Tag.Get("cli")
	if cliTag == "" {
		return
	}

	annotations := parseAnnotations(cliTag)

	// Determine if it's a group or command
	isGroup := strings.HasPrefix(annotations["group"], "group=")
	parentGroup := strings.TrimPrefix(annotations["group"], "group=")

	// Find or create parent group
	currentGroup := r.Root
	if parentGroup != "root" && parentGroup != "" {
		if g, ok := r.Root.Groups[parentGroup]; ok {
			currentGroup = g
		} else {
			// Create new group if it doesn't exist
			newGroup := &Group{
				Name:     parentGroup,
				Commands: make(map[string]Command),
				Groups:   make(map[string]*Group),
			}
			r.Root.Groups[parentGroup] = newGroup
			currentGroup = newGroup
		}
	}

	if isGroup {
		// Register as a group
		groupName := strings.TrimPrefix(annotations["group"], "group=")
		newGroup := &Group{
			Name:     groupName,
			Commands: make(map[string]Command),
			Groups:   make(map[string]*Group),
		}

		// Get help text from Help() method
		if helpMethod := v.MethodByName("Help"); helpMethod.IsValid() {
			newGroup.Help = helpMethod.Call(nil)[0].String()
		}

		currentGroup.Groups[groupName] = newGroup
	} else {
		// Register as a command
		cmd := Command{
			Execute: func(ctx Context) error {
				return v.Interface().(Runner).Run(ctx)
			},
		}

		// Get command name and help
		if cmdMethod := v.MethodByName("Command"); cmdMethod.IsValid() {
			cmd.Name = cmdMethod.Call(nil)[0].String()
		}
		if helpMethod := v.MethodByName("Help"); helpMethod.IsValid() {
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
				// Arguments are handled via reflection when running
			case strings.HasPrefix(cliTag, "option="):
				// Register option with flag
				name := strings.TrimPrefix(anns["option"], "option=")
				defaultVal := anns["default"]
				help := anns["help"]

				switch field.Type.Kind() {
				case reflect.String:
					flag.StringVar(v.Elem().Field(i).Addr().Interface().(*string), name, defaultVal, help)
				case reflect.Bool:
					boolVal := false
					if defaultVal == "true" {
						boolVal = true
					}
					flag.BoolVar(v.Elem().Field(i).Addr().Interface().(*bool), name, boolVal, help)
				case reflect.Int:
					intVal := 0
					if defaultVal != "" {
						fmt.Sscanf(defaultVal, "%d", &intVal)
					}
					flag.IntVar(v.Elem().Field(i).Addr().Interface().(*int), name, intVal, help)
				case reflect.Float64:
					floatVal := float64(0)
					if defaultVal != "" {
						fmt.Sscanf(defaultVal, "%f", &floatVal)
					}
					flag.Float64Var(v.Elem().Field(i).Addr().Interface().(*float64), name, floatVal, help)
				}
			}
		}

		currentGroup.Commands[cmd.Name] = cmd
	}
}

// Run parses arguments and executes the appropriate command
func (r *Registry) Run() {
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

	if cmd.Execute != nil {
		ctx := Context{Args: cmdArgs}
		if err := cmd.Execute(ctx); err != nil {
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
func Main(items ...interface{}) {
	registry := NewRegistry()

	// Register all items
	for _, item := range items {
		registry.Register(item)
	}

	registry.Run()
}
