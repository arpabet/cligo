/*
 * Copyright (c) 2026 Karagatan LLC.
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
	"reflect"
	"syscall"

	"go.arpabet.com/glue"
)

var RootGroup = "cli"

// parentInfo holds metadata extracted from the CliGroup parent field tag.
type parentInfo struct {
	group  string
	hidden bool
	alias  string
}

// Echo prints a formatted line to stdout. With an empty format string, it prints a blank line.
func Echo(format string, args ...interface{}) {
	if len(format) == 0 {
		println()
		return
	}
	fmt.Printf(format+"\n", args...)
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

	var beans []any

	// Resolve config files into glue PropertySource beans
	configFiles := app.getConfigFiles()
	if len(configFiles) > 0 {
		configBeans, err := resolveConfigFiles(configFiles)
		if err != nil {
			return err
		}
		beans = configBeans
	}

	beans = append(beans, app.getBeans()...)

	// Use user-provided context or create a signal-aware one
	ctx := app.getContext()
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
	}

	glueOpts := []glue.ContainerOption{glue.WithContext(ctx)}

	if profiles := app.getProfiles(); len(profiles) > 0 {
		glueOpts = append(glueOpts, glue.WithProfiles(profiles...))
	}

	if hasVerbose(os.Args[1:]) {
		glueOpts = append(glueOpts, glue.WithLogger(log.Default()))
	}

	if app.getProperties() != nil {
		glueOpts = append(glueOpts, glue.WithProperties(app.getProperties()))
	}

	glueOpts = append(glueOpts, glue.WithBeans(beans...))

	c, err := glue.NewWithOptions(glueOpts...)
	if err != nil {
		return fmt.Errorf("glue.New: %w", err)
	}
	defer c.Close()

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
