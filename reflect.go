/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

type argInfo struct {
	name     string
	position int
	required bool
	defVal   string
}

// setSliceOption sets a slice field from pflag or env var.
// For env vars, values are comma-separated (e.g. APP_TAGS=foo,bar,baz).
func (t *implCliApplication) setSliceOption(flagSet *pflag.FlagSet, f *pflag.Flag, field reflect.Value, envVars map[string]string) {
	elemKind := field.Type().Elem().Kind()

	// If flag not explicitly set, try environment variable
	if !flagSet.Changed(f.Name) {
		if envVar, ok := envVars[f.Name]; ok {
			if envValue := os.Getenv(envVar); envValue != "" {
				parts := strings.Split(envValue, ",")
				switch elemKind {
				case reflect.String:
					field.Set(reflect.ValueOf(parts))
				case reflect.Int:
					vals := make([]int, 0, len(parts))
					for _, p := range parts {
						v, _ := strconv.Atoi(strings.TrimSpace(p))
						vals = append(vals, v)
					}
					field.Set(reflect.ValueOf(vals))
				case reflect.Float64:
					vals := make([]float64, 0, len(parts))
					for _, p := range parts {
						v, _ := strconv.ParseFloat(strings.TrimSpace(p), 64)
						vals = append(vals, v)
					}
					field.Set(reflect.ValueOf(vals))
				case reflect.Bool:
					vals := make([]bool, 0, len(parts))
					for _, p := range parts {
						v, _ := strconv.ParseBool(strings.TrimSpace(p))
						vals = append(vals, v)
					}
					field.Set(reflect.ValueOf(vals))
				}
				return
			}
		}
		return
	}

	switch elemKind {
	case reflect.String:
		vals, _ := flagSet.GetStringArray(f.Name)
		field.Set(reflect.ValueOf(vals))
	case reflect.Int:
		vals, _ := flagSet.GetIntSlice(f.Name)
		field.Set(reflect.ValueOf(vals))
	case reflect.Float64:
		vals, _ := flagSet.GetFloat64Slice(f.Name)
		field.Set(reflect.ValueOf(vals))
	case reflect.Bool:
		vals, _ := flagSet.GetBoolSlice(f.Name)
		field.Set(reflect.ValueOf(vals))
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

func (t *implCliApplication) setOptionValues(flagSet *pflag.FlagSet, options map[string]reflect.Value, envVars map[string]string) {
	flagSet.VisitAll(func(f *pflag.Flag) {
		field, ok := options[f.Name]
		if !ok {
			return
		}

		if field.Kind() == reflect.Slice {
			t.setSliceOption(flagSet, f, field, envVars)
			return
		}

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
	})
}

func (t *implCliApplication) setArgumentValues(argDefs []argInfo, cmdValue reflect.Value, argValues []string, cmd CliCommand, stack []string) error {
	argIndex := 0
	for _, arg := range argDefs {
		field := cmdValue.Field(arg.position)
		if argIndex >= len(argValues) {
			if arg.required {
				Echo("%s\n%s\n", t.getCommandUsage(cmd, stack), t.getCommandTryUsage(cmd, stack))
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
				Echo("%s\n%s\n", t.getCommandUsage(cmd, stack), t.getCommandTryUsage(cmd, stack))
				return fmt.Errorf("invalid integer for argument %s: %s", arg.name, argValues[argIndex])
			}
			field.SetInt(val)
		case reflect.Float32, reflect.Float64:
			val, err := strconv.ParseFloat(argValues[argIndex], 64)
			if err != nil {
				Echo("%s\n%s\n", t.getCommandUsage(cmd, stack), t.getCommandTryUsage(cmd, stack))
				return fmt.Errorf("invalid float for argument %s: %s", arg.name, argValues[argIndex])
			}
			field.SetFloat(val)
		}
		argIndex++
	}
	return nil
}

func (t *implCliApplication) identifyArgumentsAndOptions(cmdType reflect.Type, cmdValue reflect.Value, flagSet *pflag.FlagSet) ([]argInfo, map[string]reflect.Value, map[string]string) {
	var argDefs []argInfo
	options := make(map[string]reflect.Value)
	envVars := make(map[string]string)

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
			case reflect.Slice:
				switch fieldVal.Type().Elem().Kind() {
				case reflect.String:
					if shortFlag != "" {
						flagSet.StringArrayP(optName, shortFlag, nil, helpText)
					} else {
						flagSet.StringArray(optName, nil, helpText)
					}
				case reflect.Int:
					if shortFlag != "" {
						flagSet.IntSliceP(optName, shortFlag, nil, helpText)
					} else {
						flagSet.IntSlice(optName, nil, helpText)
					}
				case reflect.Float64:
					if shortFlag != "" {
						flagSet.Float64SliceP(optName, shortFlag, nil, helpText)
					} else {
						flagSet.Float64Slice(optName, nil, helpText)
					}
				case reflect.Bool:
					if shortFlag != "" {
						flagSet.BoolSliceP(optName, shortFlag, nil, helpText)
					} else {
						flagSet.BoolSlice(optName, nil, helpText)
					}
				}
			}
		}
	}
	return argDefs, options, envVars
}
