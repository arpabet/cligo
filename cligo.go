/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

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

// implCliApplication is the main application structure
type implCliApplication struct {
	name         string
	title        string
	help         string
	version      string
	build        string
	verbose      bool
	beans        []interface{}
	properties   glue.Properties
	groups       map[string][]CliGroup
	commands     map[string][]CliCommand
	commandBeans map[string][]interface{}
	helps        map[string]string
}

// New creates a new CLI application
func New(options ...Option) CliApplication {
	app := &implCliApplication{
		groups:       make(map[string][]CliGroup),
		commands:     make(map[string][]CliCommand),
		commandBeans: make(map[string][]interface{}),
		helps:        make(map[string]string),
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

	var str strings.Builder
	if app.title != "" {
		str.WriteString(app.title)
		str.WriteString("\n")
	}
	if app.help != "" {
		str.WriteString(app.help)
		str.WriteString("\n")
	}
	app.helps[RootGroup] = str.String()

	if !app.verbose {
		app.verbose = hasVerbose(os.Args[1:])
	}

	return app
}

func (app *implCliApplication) Name() string {
	return app.name
}

func (app *implCliApplication) Title() string {
	return app.title
}

func (app *implCliApplication) Help() string {
	return app.help
}

func (app *implCliApplication) Version() string {
	return app.version
}

func (app *implCliApplication) Build() string {
	return app.build
}

func (app *implCliApplication) Verbose() bool {
	return app.verbose
}

func (app *implCliApplication) getBeans() []interface{} {
	return app.beans
}

func (app *implCliApplication) getProperties() glue.Properties {
	return app.properties
}

func hasVerbose(args []string) bool {
	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			return true
		}
	}
	return false
}

func Echo(format string, args ...interface{}) {
	if len(format) == 0 {
		println()
		return
	}
	fmt.Printf(format+"\n", args...)
}

// RegisterGroup registers a command group
func (app *implCliApplication) RegisterGroup(group CliGroup) error {
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
func (app *implCliApplication) RegisterCommand(cmd CliCommand) error {
	parentGroup := extractParentGroup(cmd)
	if parentGroup == "" {
		return errors.Errorf("parent group not found in cli command: %v", cmd)
	}
	app.commands[parentGroup] = append(app.commands[parentGroup], cmd)
	return nil
}

// RegisterCommandWithBeans registers a command with beans
func (app *implCliApplication) RegisterCommandWithBeans(cmd CliCommandWithBeans) error {
	parentGroup := extractParentGroup(cmd)
	if parentGroup == "" {
		return errors.Errorf("parent group not found in cli command: %v", cmd)
	}
	app.commands[parentGroup] = append(app.commands[parentGroup], cmd)

	commandBeans := cmd.CommandBeans()
	if len(commandBeans) > 0 {
		app.commandBeans[cmd.Command()] = append(app.commandBeans[cmd.Command()], commandBeans...)
	}
	return nil
}

// Execute parses arguments and runs the appropriate command
func (app *implCliApplication) Execute(ctx glue.Context) error {

	if len(os.Args) < 2 {
		app.printHelp(RootGroup, nil)
		return nil
	}

	// Check for version flag
	if app.version != "" {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			name := app.name
			if app.title != "" {
				name = app.title
			}
			if app.build != "" {
				Echo("%s Version %s Build %s", name, app.version, app.build)
			} else {
				Echo("%s Version %s", name, app.version)
			}
			if app.help != "" {
				Echo(app.help)
			}
			return nil
		}
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
func (app *implCliApplication) parseAndExecute(ctx glue.Context, currentGroup string, args []string, stack []string) error {
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

	// Check if the first argument is a know option
	if args[0] == "--help" || args[0] == "-h" {
		app.printHelp(RootGroup, stack)
		return nil
	}

	if args[0] == "--verbose" || args[0] == "-v" {
		app.verbose = true
		app.printHelp(currentGroup, stack)
		return nil
	}

	app.printHelp(currentGroup, stack)
	return fmt.Errorf("unknown command or group: %s", args[0])
}

// executeCommand parses arguments and options for a command and executes it
func (app *implCliApplication) executeCommand(ctx glue.Context, cmd CliCommand, args []string, stack []string) error {
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
	isVerbose := flagSet.Bool("verbose", false, "Verbose output")

	// Parse flags
	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	argValues := flagSet.Args()

	if *isHelp {
		app.printCommandHelp(cmd, stack)
		return nil
	}

	// update verbose flag based on options
	app.verbose = *isVerbose

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

	cmdBeans, ok := app.commandBeans[cmd.Command()]
	if ok && len(cmdBeans) > 0 {
		child, err := ctx.Extend(cmdBeans...)
		if err != nil {
			Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
			return fmt.Errorf("fail to initialize '%s' command scope context, %v", cmd.Command(), err)
		}
		defer child.Close()
		return cmd.Run(child)
	}

	// Execute the command in the appication context
	return cmd.Run(ctx)
}

// printHelp prints help for a group
func (app *implCliApplication) printHelp(groupName string, stack []string) {

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
		if app.version != "" {
			Echo("  --version  Show the version and exit.")
		}
		Echo("  --verbose  Show extended logging information.")
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
func (app *implCliApplication) getCommandUsage(cmd CliCommand, stack []string) string {

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

	return fmt.Sprintf("Usage: %s %s [OPTIONS] %s", app.name, path, argsLine)
}

// getCommandTryUsage gets printable help with try statement
func (app *implCliApplication) getCommandTryUsage(cmd CliCommand, stack []string) string {
	path := strings.Join(stack, " ")
	return fmt.Sprintf("Try '%s %s --help' for help", app.name, path)
}

// printCommandHelp prints help for a specific command
func (app *implCliApplication) printCommandHelp(cmd CliCommand, stack []string) {

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

// Run entry point
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

	var ctx glue.Context
	if app.getProperties() != nil {
		ctx, err = glue.NewWithProperties(app.getProperties(), app.getBeans()...)
	} else {
		ctx, err = glue.New(app.getBeans()...)
	}
	if err != nil {
		return errors.Errorf("glue.New: %v", err)
	}
	defer ctx.Close()

	visited := make(map[uintptr]bool)

	// Register all groups
	for _, item := range ctx.Bean(CliGroupClass, 0) {
		obj := item.Object()
		addr := reflect.ValueOf(obj).Pointer()
		if visited[addr] {
			continue
		}
		visited[addr] = true
		err = app.RegisterGroup(obj.(CliGroup))
		if err != nil {
			return err
		}
	}

	// Register all commands with beans
	for _, item := range ctx.Bean(CliCommandWithBeansClass, 0) {
		obj := item.Object()
		addr := reflect.ValueOf(obj).Pointer()
		if visited[addr] {
			continue
		}
		visited[addr] = true
		err = app.RegisterCommandWithBeans(obj.(CliCommandWithBeans))
		if err != nil {
			return err
		}
	}

	// Register all commands
	for _, item := range ctx.Bean(CliCommandClass, 0) {
		obj := item.Object()
		addr := reflect.ValueOf(obj).Pointer()
		if visited[addr] {
			continue
		}
		visited[addr] = true
		err = app.RegisterCommand(obj.(CliCommand))
		if err != nil {
			return err
		}
	}

	return app.Execute(ctx)
}

func Main(options ...Option) {

	if err := Run(options...); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
