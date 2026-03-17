/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

// Package cligo is a declarative CLI framework for Go, inspired by Python's Click.
// Commands and groups are defined as structs implementing CliCommand and CliGroup interfaces,
// with arguments and options declared via struct tags. Built on top of the glue DI framework.
package cligo

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	"go.arpabet.com/glue"
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
	color        *bool
	ctx          context.Context
	beans        []interface{}
	properties   glue.Properties
	groups       map[string][]CliGroup
	commands     map[string][]CliCommand
	commandBeans map[string][]interface{}
	helps        map[string]string
	hidden       map[interface{}]bool
	aliasOf      map[interface{}]string
	cmdAliases   map[string]map[string]CliCommand
	groupAliases map[string]map[string]CliGroup
}

// parentInfo holds metadata extracted from the CliGroup parent field tag.
type parentInfo struct {
	group  string
	hidden bool
	alias  string
}

// New creates a new CLI application
func New(options ...Option) CliApplication {
	app := &implCliApplication{
		groups:       make(map[string][]CliGroup),
		commands:     make(map[string][]CliCommand),
		commandBeans: make(map[string][]interface{}),
		helps:        make(map[string]string),
		hidden:       make(map[interface{}]bool),
		aliasOf:      make(map[interface{}]string),
		cmdAliases:   make(map[string]map[string]CliCommand),
		groupAliases: make(map[string]map[string]CliGroup),
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

func (app *implCliApplication) getContext() context.Context {
	return app.ctx
}

func hasVerbose(args []string) bool {
	for _, arg := range args {
		if arg == "--verbose" {
			return true
		}
	}
	return false
}

// Echo prints a formatted line to stdout. With an empty format string, it prints a blank line.
func Echo(format string, args ...interface{}) {
	if len(format) == 0 {
		println()
		return
	}
	fmt.Printf(format+"\n", args...)
}

// RegisterGroup registers a command group
func (app *implCliApplication) RegisterGroup(group CliGroup) error {
	info := extractParentInfo(group)
	if info.group == "" {
		return fmt.Errorf("parent group not found in cli group: %v", group)
	}
	app.groups[info.group] = append(app.groups[info.group], group)
	if info.hidden {
		app.hidden[group] = true
	}
	if info.alias != "" {
		app.aliasOf[group] = info.alias
		if app.groupAliases[info.group] == nil {
			app.groupAliases[info.group] = make(map[string]CliGroup)
		}
		app.groupAliases[info.group][info.alias] = group
	}
	shortDesc, longDesc := group.Help()
	if len(longDesc) == 0 {
		longDesc = shortDesc
	}
	app.helps[group.Group()] = longDesc
	return nil
}

// RegisterCommand registers a command
func (app *implCliApplication) RegisterCommand(cmd CliCommand) error {
	info := extractParentInfo(cmd)
	if info.group == "" {
		return fmt.Errorf("parent group not found in cli command: %v", cmd)
	}
	app.commands[info.group] = append(app.commands[info.group], cmd)
	if info.hidden {
		app.hidden[cmd] = true
	}
	if info.alias != "" {
		app.aliasOf[cmd] = info.alias
		if app.cmdAliases[info.group] == nil {
			app.cmdAliases[info.group] = make(map[string]CliCommand)
		}
		app.cmdAliases[info.group][info.alias] = cmd
	}
	return nil
}

// RegisterCommandWithBeans registers a command with beans
func (app *implCliApplication) RegisterCommandWithBeans(cmd CliCommandWithBeans) error {
	info := extractParentInfo(cmd)
	if info.group == "" {
		return fmt.Errorf("parent group not found in cli command: %v", cmd)
	}
	app.commands[info.group] = append(app.commands[info.group], cmd)
	if info.hidden {
		app.hidden[cmd] = true
	}
	if info.alias != "" {
		app.aliasOf[cmd] = info.alias
		if app.cmdAliases[info.group] == nil {
			app.cmdAliases[info.group] = make(map[string]CliCommand)
		}
		app.cmdAliases[info.group][info.alias] = cmd
	}

	commandBeans := cmd.CommandBeans()
	if len(commandBeans) > 0 {
		app.commandBeans[cmd.Command()] = append(app.commandBeans[cmd.Command()], commandBeans...)
	}
	return nil
}

// Execute parses arguments and runs the appropriate command
func (app *implCliApplication) Execute(ctx context.Context, c glue.Container) error {

	if len(os.Args) < 2 {
		app.printHelp(RootGroup, nil)
		return nil
	}

	// Check for version flag
	if app.version != "" {
		if os.Args[1] == "--version" || os.Args[1] == "-V" {
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
	return app.parseAndExecute(ctx, c, RootGroup, os.Args[1:], stack)
}

// parseAndExecute recursively parses arguments and executes the appropriate command
func (app *implCliApplication) parseAndExecute(ctx context.Context, c glue.Container, currentGroup string, args []string, stack []string) error {
	if len(args) == 0 {
		app.printHelp(currentGroup, stack)
		return nil
	}

	// Check if the first argument is a group (by name or alias)
	matchedGroup := app.findGroup(currentGroup, args[0])
	if matchedGroup != nil {
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			app.printHelp(matchedGroup.Group(), stack)
			return nil
		}
		stack = append(stack, args[0])
		return app.parseAndExecute(ctx, c, matchedGroup.Group(), args[1:], stack)
	}

	// Check if the first argument is a command (by name or alias)
	matchedCmd := app.findCommand(currentGroup, args[0])
	if matchedCmd != nil {
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			app.printCommandHelp(matchedCmd, stack)
			return nil
		}
		stack = append(stack, args[0])
		return app.executeCommand(ctx, c, matchedCmd, args[1:], stack)
	}

	// Check if the first argument is a know option
	if args[0] == "--help" || args[0] == "-h" {
		app.printHelp(RootGroup, stack)
		return nil
	}

	if args[0] == "--verbose" {
		app.verbose = true
		app.printHelp(currentGroup, stack)
		return nil
	}

	app.printHelp(currentGroup, stack)
	return fmt.Errorf("unknown command or group: %s", args[0])
}

func (app *implCliApplication) findGroup(parentGroup, name string) CliGroup {
	for _, group := range app.groups[parentGroup] {
		if group.Group() == name {
			return group
		}
	}
	if aliasMap, ok := app.groupAliases[parentGroup]; ok {
		if group, ok := aliasMap[name]; ok {
			return group
		}
	}
	return nil
}

func (app *implCliApplication) findCommand(parentGroup, name string) CliCommand {
	for _, cmd := range app.commands[parentGroup] {
		if cmd.Command() == name {
			return cmd
		}
	}
	if aliasMap, ok := app.cmdAliases[parentGroup]; ok {
		if cmd, ok := aliasMap[name]; ok {
			return cmd
		}
	}
	return nil
}

// executeCommand parses arguments and options for a command and executes it
func (app *implCliApplication) executeCommand(ctx context.Context, c glue.Container, cmd CliCommand, args []string, stack []string) error {
	// Create a new value to store the parsed arguments
	cmdValue := reflect.ValueOf(cmd).Elem()
	cmdType := cmdValue.Type()

	// Prepare a custom flag set
	flagSet := pflag.NewFlagSet(cmd.Command(), pflag.ContinueOnError)
	flagSet.Usage = func() { app.printCommandHelp(cmd, stack) }

	type argInfo struct {
		name     string
		position int
		required bool
		defVal   string
	}

	var argDefs []argInfo
	options := make(map[string]reflect.Value)
	envVars := make(map[string]string) // option name -> env var name

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
			_, hasDefault := tagParts["default"]
			_, hasRequired := tagParts["required"]
			argDefs = append(argDefs, argInfo{
				name:     argName,
				position: i,
				required: !hasDefault || hasRequired,
				defVal:   tagParts["default"],
			})
			continue
		}

		// Handle option
		if optName, ok := tagParts["option"]; ok {
			fieldVal := cmdValue.Field(i)
			options[optName] = fieldVal

			shortFlag := strings.TrimPrefix(tagParts["short"], "-")
			helpText := tagParts["help"]

			// Track environment variable binding
			if envVar, ok := tagParts["env"]; ok {
				envVars[optName] = envVar
				if helpText != "" {
					helpText = helpText + " [$" + envVar + "]"
				} else {
					helpText = "[$" + envVar + "]"
				}
			}

			// Register flag with the flag set based on field type
			switch fieldVal.Kind() {
			case reflect.String:
				defaultVal := tagParts["default"]
				if shortFlag != "" {
					flagSet.StringP(optName, shortFlag, defaultVal, helpText)
				} else {
					flagSet.String(optName, defaultVal, helpText)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				defaultVal := 0
				if val, ok := tagParts["default"]; ok {
					defaultVal, _ = strconv.Atoi(val)
				}
				if shortFlag != "" {
					flagSet.IntP(optName, shortFlag, defaultVal, helpText)
				} else {
					flagSet.Int(optName, defaultVal, helpText)
				}
			case reflect.Float32, reflect.Float64:
				defaultVal := 0.0
				if val, ok := tagParts["default"]; ok {
					defaultVal, _ = strconv.ParseFloat(val, 64)
				}
				if shortFlag != "" {
					flagSet.Float64P(optName, shortFlag, defaultVal, helpText)
				} else {
					flagSet.Float64(optName, defaultVal, helpText)
				}
			case reflect.Bool:
				defaultVal := false
				if val, ok := tagParts["default"]; ok {
					defaultVal = val == "true"
				}
				if shortFlag != "" {
					flagSet.BoolP(optName, shortFlag, defaultVal, helpText)
				} else {
					flagSet.Bool(optName, defaultVal, helpText)
				}
			}
		}
	}

	// Add help option
	isHelp := flagSet.BoolP("help", "h", false, "Print help")
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

	// update verbose flag only if explicitly passed at command level
	if *isVerbose {
		app.verbose = true
	}

	// Set argument values
	argIndex := 0
	for _, arg := range argDefs {
		field := cmdValue.Field(arg.position)
		if argIndex >= len(argValues) {
			if arg.required {
				Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
				return fmt.Errorf("missing required argument '%s'", arg.name)
			}
			if arg.defVal != "" {
				setFieldFromString(field, arg.defVal)
			}
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(argValues[argIndex])
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, err := strconv.ParseInt(argValues[argIndex], 10, 64)
			if err != nil {
				Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
				return fmt.Errorf("invalid integer for argument %s: %s", arg.name, argValues[argIndex])
			}
			field.SetInt(val)
		case reflect.Float32, reflect.Float64:
			val, err := strconv.ParseFloat(argValues[argIndex], 64)
			if err != nil {
				Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
				return fmt.Errorf("invalid float for argument %s: %s", arg.name, argValues[argIndex])
			}
			field.SetFloat(val)
		}
		argIndex++
	}

	// Set option values: explicit flag > env var > default.
	flagSet.VisitAll(func(f *pflag.Flag) {
		if field, ok := options[f.Name]; ok {
			value := f.Value.String()

			// If flag not explicitly set, try environment variable
			if !flagSet.Changed(f.Name) {
				if envVar, ok := envVars[f.Name]; ok {
					if envValue := os.Getenv(envVar); envValue != "" {
						value = envValue
					}
				}
			}

			setFieldFromString(field, value)
		}
	})

	cmdBeans, ok := app.commandBeans[cmd.Command()]
	if ok && len(cmdBeans) > 0 {
		child, err := c.Extend(cmdBeans...)
		if err != nil {
			Echo("%s\n%s\n", app.getCommandUsage(cmd, stack), app.getCommandTryUsage(cmd, stack))
			return fmt.Errorf("fail to initialize '%s' command scope context, %v", cmd.Command(), err)
		}
		defer child.Close()
		return cmd.Run(ctx, child)
	}

	// Execute the command in the application context
	return cmd.Run(ctx, c)
}

// printHelp prints help for a group
func (app *implCliApplication) printHelp(groupName string, stack []string) {

	groups := app.groups[groupName]
	commands := app.commands[groupName]

	path := strings.Join(stack, " ")

	if len(groups)+len(commands) > 0 {
		Echo("%s: %s %s [OPTIONS] COMMAND [ARGS]...", app.styled("Usage", ansiBold), app.name, path)
	} else {
		Echo("%s: %s %s [OPTIONS] [ARGS]...", app.styled("Usage", ansiBold), app.name, path)
	}

	help := app.helps[groupName]
	if help != "" {
		Echo("\n%s\n", help)
	}

	if groupName == RootGroup {
		Echo("%s:", app.styled("Options", ansiBold))
		if app.version != "" {
			Echo("  %s  Show the version and exit.", app.styled("--version", ansiYellow))
		}
		Echo("  %s  Show extended logging information.", app.styled("--verbose", ansiYellow))
		Echo("  %s     Show this message and exit.", app.styled("--help", ansiYellow))
		Echo("")
	}

	Echo("%s:", app.styled("Commands", ansiBold))
	for _, grp := range groups {
		if app.hidden[grp] {
			continue
		}
		shortDesc, _ := grp.Help()
		name := app.styled(grp.Group(), ansiCyan)
		if alias, ok := app.aliasOf[grp]; ok {
			name = name + " (" + alias + ")"
		}
		Echo("  %s\t%s", name, shortDesc)
	}

	for _, cmd := range commands {
		if app.hidden[cmd] {
			continue
		}
		shortDesc, _ := cmd.Help()
		name := app.styled(cmd.Command(), ansiCyan)
		if alias, ok := app.aliasOf[cmd]; ok {
			name = name + " (" + alias + ")"
		}
		Echo("  %s\t%s", name, shortDesc)
	}

}

// getCommandUsage gets printable usage line
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
			name := strings.ToUpper(argName)
			_, hasDefault := tagParts["default"]
			_, hasRequired := tagParts["required"]
			if hasDefault && !hasRequired {
				name = "[" + name + "]"
			}
			arguments = append(arguments, name)
		}
	}

	path := strings.Join(stack, " ")
	argsLine := strings.Join(arguments, " ")

	return fmt.Sprintf("%s: %s %s [OPTIONS] %s", app.styled("Usage", ansiBold), app.name, path, argsLine)
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

	hasOptions := false

	Echo(app.getCommandUsage(cmd, stack))

	shortDesc, longDesc := cmd.Help()
	if len(longDesc) == 0 {
		longDesc = shortDesc
	}

	Echo("%s\n", longDesc)

	// Print argument details
	var argLines []string
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
			_, hasDefault := tagParts["default"]
			_, hasRequired := tagParts["required"]
			if hasDefault && !hasRequired {
				help = help + fmt.Sprintf(" [default: %s]", tagParts["default"])
			} else {
				help = help + " [required]"
			}
			argLines = append(argLines, fmt.Sprintf("  %s\t%s", app.styled(strings.ToUpper(argName), ansiGreen), help))
		}
	}
	if len(argLines) > 0 {
		Echo("%s:", app.styled("Arguments", ansiBold))
		for _, line := range argLines {
			fmt.Println(line)
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
				Echo("%s:", app.styled("Options", ansiBold))
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

			envText := ""
			if envVar, ok := tagParts["env"]; ok {
				envText = fmt.Sprintf(" [$%s]", envVar)
			}

			fmt.Printf("  %s  %s%s%s\n", app.styled("--"+optName, ansiYellow), help, defaultText, envText)
		}
	}
}

// setFieldFromString sets a reflect.Value from a string, handling type conversion.
func setFieldFromString(field reflect.Value, value string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, _ := strconv.ParseInt(value, 10, 64)
		field.SetInt(val)
	case reflect.Float32, reflect.Float64:
		val, _ := strconv.ParseFloat(value, 64)
		field.SetFloat(val)
	case reflect.Bool:
		val, _ := strconv.ParseBool(value)
		field.SetBool(val)
	}
}

// parseCliTag parses a cli tag string into a map of key-value pairs
func parseCliTag(tag string) map[string]string {
	result := make(map[string]string)
	if tag == "" {
		return result
	}
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

// extractParentInfo extracts group, hidden, and alias metadata from the CliGroup parent field.
func extractParentInfo(obj interface{}) parentInfo {
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Type == CliGroupClass {
			cliTag := field.Tag.Get("cli")
			if cliTag != "" {
				tagParts := parseCliTag(cliTag)
				_, isHidden := tagParts["hidden"]
				return parentInfo{
					group:  tagParts["group"],
					hidden: isHidden,
					alias:  tagParts["alias"],
				}
			}
		}
	}

	return parentInfo{}
}

// extractParentGroup extracts the parent group name from a command or group.
func extractParentGroup(obj interface{}) string {
	return extractParentInfo(obj).group
}

// Run creates the application, sets up the glue DI container, discovers all
// registered groups and commands, then parses os.Args and executes the matched command.
// Returns an error on failure. Panics from command execution are recovered and returned as errors.
func Run(options ...Option) (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = fmt.Errorf("%s", v)
			default:
				err = fmt.Errorf("recover: %v", v)
			}
		}
	}()

	app := New(options...)

	var glueOpts []glue.ContainerOption

	if hasVerbose(os.Args[1:]) {
		glueOpts = append(glueOpts, glue.WithLogger(log.Default()))
	}

	if app.getProperties() != nil {
		glueOpts = append(glueOpts, glue.WithProperties(app.getProperties()))
	}

	c, err := glue.NewWithOptions(glueOpts, app.getBeans()...)
	if err != nil {
		return fmt.Errorf("glue.New: %w", err)
	}
	defer c.Close()

	// Use user-provided context or create a signal-aware one
	ctx := app.getContext()
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
	}

	visited := make(map[uintptr]bool)

	// Register all groups
	for _, item := range c.Bean(CliGroupClass, 0) {
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
	for _, item := range c.Bean(CliCommandWithBeansClass, 0) {
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
	for _, item := range c.Bean(CliCommandClass, 0) {
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

	return app.Execute(ctx, c)
}

// Main is the standard entry point for CLI applications.
// It calls Run and prints the error to stdout and exits with code 1 on failure.
func Main(options ...Option) {

	if err := Run(options...); err != nil {
		// Detect color for error output without needing the app instance
		errPrefix := "Error"
		fi, statErr := os.Stderr.Stat()
		if statErr == nil && fi.Mode()&os.ModeCharDevice != 0 && os.Getenv("NO_COLOR") == "" {
			errPrefix = ansiRed + ansiBold + "Error" + ansiReset
		}
		fmt.Fprintf(os.Stderr, "%s: %v\n", errPrefix, err)
		os.Exit(1)
	}
}
