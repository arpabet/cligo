package cligo

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"go.arpabet.com/glue"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

var RootGroup = "cli"

// App is the main application structure
type App struct {
	name     string
	help     string
	version  string
	build    string
	beans    []interface{}
	groups   map[string][]CliGroup
	commands map[string][]CliCommand
	helps    map[string]string
}

// New creates a new CLI application
func New(options ...Option) *App {
	app := &App{
		groups:   make(map[string][]CliGroup),
		commands: make(map[string][]CliCommand),
		helps:    make(map[string]string),
	}

	// first bean is application itself
	app.beans = []interface{}{app}

	// apply options
	for _, opt := range options {
		opt.apply(app)
	}

	if app.name == "" {
		app.name = filepath.Base(os.Args[0])
	}

	if app.help != "" {
		app.helps[RootGroup] = app.help
	}

	return app
}

func Echo(format string, args ...interface{}) {
	if len(format) == 0 {
		println()
		return
	}
	fmt.Printf(format+"\n", args...)
}

// RegisterGroup registers a command group
func (app *App) RegisterGroup(group CliGroup) error {
	parentGroup := extractParentGroup(group)
	if parentGroup == "" {
		return errors.Errorf("parent group not found in cli group: %v", group)
	}
	app.groups[parentGroup] = append(app.groups[parentGroup], group)
	shortDesc, longDesc := group.Help()
	if len(longDesc) == 0 {
		longDesc = shortDesc
	}
	app.helps[group.Group()] = longDesc
	return nil
}

// RegisterCommand registers a command
func (app *App) RegisterCommand(cmd CliCommand) error {
	parentGroup := extractParentGroup(cmd)
	if parentGroup == "" {
		return errors.Errorf("parent group not found in cli command: %v", cmd)
	}
	app.commands[parentGroup] = append(app.commands[parentGroup], cmd)
	return nil
}

// RunCLI parses arguments and runs the appropriate command
func (app *App) RunCLI(ctx glue.Context) error {

	if len(os.Args) < 2 {
		app.printHelp(RootGroup, nil)
		return nil
	}

	// Check for version flag
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		Echo("Application: %s", app.name)
		if app.version != "" {
			Echo("Version: %s", app.version)
		}
		if app.build != "" {
			Echo("Build: %s", app.build)
		}
		return nil
	}

	// Check for help flag
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		app.printHelp(RootGroup, nil)
		return nil
	}

	var stack []string
	return app.parseAndExecute(ctx, RootGroup, os.Args[1:], stack)
}

// parseAndExecute recursively parses arguments and executes the appropriate command
func (app *App) parseAndExecute(ctx glue.Context, currentGroup string, args []string, stack []string) error {
	if len(args) == 0 {
		app.printHelp(currentGroup, stack)
		return nil
	}

	// Check if the first argument is a group
	for _, group := range app.groups[currentGroup] {
		if group.Group() == args[0] {
			if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
				app.printHelp(group.Group(), stack)
				return nil
			}
			stack = append(stack, args[0])
			return app.parseAndExecute(ctx, group.Group(), args[1:], stack)
		}
	}

	// Check if the first argument is a command
	for _, cmd := range app.commands[currentGroup] {
		if cmd.Command() == args[0] {
			if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
				app.printCommandHelp(cmd, stack)
				return nil
			}
			stack = append(stack, args[0])
			return app.executeCommand(ctx, cmd, args[1:], stack)
		}
	}

	fmt.Printf("Unknown command or group: %s\n", args[0])
	app.printHelp(currentGroup, stack)
	return fmt.Errorf("unknown command or group: %s", args[0])
}

// executeCommand parses arguments and options for a command and executes it
func (app *App) executeCommand(ctx glue.Context, cmd CliCommand, args []string, stack []string) error {
	// Create a new value to store the parsed arguments
	cmdValue := reflect.ValueOf(cmd).Elem()
	cmdType := cmdValue.Type()

	// Prepare a custom flag set
	flagSet := pflag.NewFlagSet(cmd.Command(), pflag.ContinueOnError)
	flagSet.Usage = func() { app.printCommandHelp(cmd, stack) }

	// Track arguments and their positions
	var arguments []string
	var positions []int
	options := make(map[string]reflect.Value)

	// First pass: identify arguments and register options
	for i := 0; i < cmdType.NumField(); i++ {
		field := cmdType.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue
		}

		tagParts := parseCliTag(cliTag)

		// Handle argument
		if argName, ok := tagParts["argument"]; ok {
			arguments = append(arguments, argName)
			positions = append(positions, i)
			continue
		}

		// Handle option
		if optName, ok := tagParts["option"]; ok {
			fieldVal := cmdValue.Field(i)
			options[optName] = fieldVal

			// Register flag with the flag set based on field type
			switch fieldVal.Kind() {
			case reflect.String:
				defaultVal := tagParts["default"]
				helpText := tagParts["help"]
				flagSet.String(optName, defaultVal, helpText)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				defaultVal := 0
				if val, ok := tagParts["default"]; ok {
					defaultVal, _ = strconv.Atoi(val)
				}
				helpText := tagParts["help"]
				flagSet.Int(optName, defaultVal, helpText)
			case reflect.Float32, reflect.Float64:
				defaultVal := 0.0
				if val, ok := tagParts["default"]; ok {
					defaultVal, _ = strconv.ParseFloat(val, 64)
				}
				helpText := tagParts["help"]
				flagSet.Float64(optName, defaultVal, helpText)
			case reflect.Bool:
				defaultVal := false
				if val, ok := tagParts["default"]; ok {
					defaultVal = val == "true"
				}
				helpText := tagParts["help"]
				flagSet.Bool(optName, defaultVal, helpText)
			}
		}
	}

	// Add help option
	isHelp := flagSet.Bool("help", false, "Print help")

	// Parse flags
	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	argValues := flagSet.Args()

	print("argValues", strings.Join(argValues, ","))

	if *isHelp {
		app.printCommandHelp(cmd, stack)
		return nil
	}

	// Handle positional arguments
	//if len(argValues) < len(arguments) {
	//	Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
	//	return fmt.Errorf("not enough arguments provided, expected %d, got %d", len(arguments), len(argValues))
	//}

	// Set argument values
	argIndex := 0
	for i, argName := range arguments {
		fieldIndex := positions[i]
		field := cmdValue.Field(fieldIndex)
		if argIndex >= len(argValues) {
			Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
			return fmt.Errorf("missing argument '%s' on position %d", argName, fieldIndex)
		}

		// Set the field value based on its type
		switch field.Kind() {
		case reflect.String:
			field.SetString(argValues[argIndex])
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, err := strconv.ParseInt(argValues[argIndex], 10, 64)
			if err != nil {
				Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
				return fmt.Errorf("invalid integer for argument %s: %s", argName, argValues[argIndex])
			}
			field.SetInt(val)
		case reflect.Float32, reflect.Float64:
			val, err := strconv.ParseFloat(argValues[argIndex], 64)
			if err != nil {
				Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
				return fmt.Errorf("invalid float for argument %s: %s", argName, argValues[argIndex])
			}
			field.SetFloat(val)
		}
		argIndex++
	}

	// Set option values
	flagSet.Visit(func(f *pflag.Flag) {
		if field, ok := options[f.Name]; ok {
			// Set the field value based on its type
			switch field.Kind() {
			case reflect.String:
				field.SetString(f.Value.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val, _ := strconv.ParseInt(f.Value.String(), 10, 64)
				field.SetInt(val)
			case reflect.Float32, reflect.Float64:
				val, _ := strconv.ParseFloat(f.Value.String(), 64)
				field.SetFloat(val)
			case reflect.Bool:
				val, _ := strconv.ParseBool(f.Value.String())
				field.SetBool(val)
			}
		}
	})

	// Execute the command
	return cmd.Run(ctx)
}

// printHelp prints help for a group
func (app *App) printHelp(groupName string, stack []string) {

	groups := app.groups[groupName]
	commands := app.commands[groupName]

	path := strings.Join(stack, " ")

	if len(groups)+len(commands) > 0 {
		Echo("Usage: %s %s [OPTIONS] COMMAND [ARGS]...", app.name, path)
	} else {
		Echo("Usage: %s %s [OPTIONS] [ARGS]...", app.name, path)
	}

	help := app.helps[groupName]
	if help != "" {
		Echo("\n%s\n", help)
	}

	if groupName == RootGroup {
		Echo("Options:")
		Echo("  --version  Show the version and exit.")
		Echo("  --help     Show this message and exit.")
		Echo("")
	}

	Echo("Commands:")
	for _, grp := range groups {
		shortDesc, _ := grp.Help()
		Echo("  %s\t%s", grp.Group(), shortDesc)
	}

	for _, cmd := range commands {
		shortDesc, _ := cmd.Help()
		Echo("  %s\t%s", cmd.Command(), shortDesc)
	}

}

// getCommandTryUsage gets printable help
func (app *App) getCommandUsage(cmd CliCommand, stack []string) string {

	// Print arguments and options
	cmdValue := reflect.ValueOf(cmd).Elem()
	cmdType := cmdValue.Type()

	// First get arguments
	var arguments []string
	for i := 0; i < cmdType.NumField(); i++ {
		field := cmdType.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue
		}

		tagParts := parseCliTag(cliTag)
		if argName, ok := tagParts["argument"]; ok {
			arguments = append(arguments, strings.ToUpper(argName))
		}
	}

	path := strings.Join(stack, " ")
	argsLine := strings.Join(arguments, " ")

	return fmt.Sprintf("Usage: %s %s [OPTIONS] %s", cmd.Command(), path, argsLine)
}

// getCommandTryUsage gets printable help with try statement
func (app *App) getCommandTryUsage(cmd CliCommand, stack []string) string {
	path := strings.Join(stack, " ")
	return fmt.Sprintf("Try '%s %s --help' for help", cmd.Command(), path)
}

// printCommandHelp prints help for a specific command
func (app *App) printCommandHelp(cmd CliCommand, stack []string) {

	// Print arguments and options
	cmdValue := reflect.ValueOf(cmd).Elem()
	cmdType := cmdValue.Type()

	hasArgs := false
	hasOptions := false

	Echo(app.getCommandUsage(cmd, stack))

	shortDesc, longDesc := cmd.Help()
	if len(longDesc) == 0 {
		longDesc = shortDesc
	}

	Echo("%s\n", longDesc)

	// Then print argument details
	if hasArgs {
		fmt.Println("Arguments:")
		for i := 0; i < cmdType.NumField(); i++ {
			field := cmdType.Field(i)
			cliTag := field.Tag.Get("cli")
			if cliTag == "" {
				continue
			}

			tagParts := parseCliTag(cliTag)
			if argName, ok := tagParts["argument"]; ok {
				help := tagParts["help"]
				if help == "" {
					help = fmt.Sprintf("%s argument", argName)
				}
				fmt.Printf("  %s => %s\n", strings.ToUpper(argName), help)
			}
		}
		fmt.Println()
	}

	// Finally print option details
	for i := 0; i < cmdType.NumField(); i++ {
		field := cmdType.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue
		}

		tagParts := parseCliTag(cliTag)
		if optName, ok := tagParts["option"]; ok {
			if !hasOptions {
				fmt.Println("Options:")
				hasOptions = true
			}

			defaultVal := tagParts["default"]
			help := tagParts["help"]
			if help == "" {
				help = fmt.Sprintf("%s option", optName)
			}

			defaultText := ""
			if defaultVal != "" {
				defaultText = fmt.Sprintf(" [default: %s]", defaultVal)
			}

			fmt.Printf("  --%s  %s%s\n", optName, help, defaultText)
		}
	}
}

// parseCliTag parses a cli tag string into a map of key-value pairs
func parseCliTag(tag string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else if len(kv) == 1 {
			// Handle boolean flags or other special cases
			result[kv[0]] = "true"
		}
	}

	return result
}

// extractParentGroup extracts the parent group from a command or group
func extractParentGroup(obj interface{}) string {
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Look for the 'CliGroupClass' field with cli tag
		if field.Type == CliGroupClass {
			cliTag := field.Tag.Get("cli")
			if cliTag != "" {
				tagParts := parseCliTag(cliTag)
				if groupName, ok := tagParts["group"]; ok {
					return groupName
				}
			}
		}
	}

	return ""
}

// Main entry point
func Run(options ...Option) (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = errors.Errorf("recover:  %v", v)
			}
		}
	}()

	app := New(options...)

	if hasVerbose(os.Args[1:]) {
		glue.Verbose(log.Default())
	}

	ctx, err := glue.New(app.beans...)
	if err != nil {
		return errors.Errorf("glue.New: %v", err)
	}
	defer ctx.Close()

	// Register all groups
	for _, item := range ctx.Bean(CliGroupClass, 0) {
		err = app.RegisterGroup(item.Object().(CliGroup))
		if err != nil {
			return err
		}
	}

	// Register all commands
	for _, item := range ctx.Bean(CliCommandClass, 0) {
		err = app.RegisterCommand(item.Object().(CliCommand))
		if err != nil {
			return err
		}
	}

	return app.RunCLI(ctx)
}

func hasVerbose(args []string) bool {
	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			return true
		}
	}
	return false
}

func Main(options ...Option) {

	if err := Run(options...); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
