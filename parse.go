/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "strings"

// parseGlobalFlag extracts all values for a given --flag/-short from args.
// Supports --flag value, --flag=value, -s value, -s=value, and repeated usage.
func parseGlobalFlag(args []string, flag string, short string) []string {
	longPrefix := "--" + flag
	shortPrefix := "-" + short
	var values []string
	for i, arg := range args {
		var value string
		if (arg == longPrefix || arg == shortPrefix) && i+1 < len(args) {
			value = args[i+1]
		} else if strings.HasPrefix(arg, longPrefix+"=") {
			value = strings.TrimPrefix(arg, longPrefix+"=")
		} else if strings.HasPrefix(arg, shortPrefix+"=") {
			value = strings.TrimPrefix(arg, shortPrefix+"=")
		}
		if value != "" {
			for _, v := range strings.Split(value, ",") {
				v = strings.TrimSpace(v)
				if v != "" {
					values = append(values, v)
				}
			}
		}
	}
	return values
}

// parseGlobalProperties extracts -D/--property key=value overrides from args.
// Accepted forms (all repeatable): -Dkey=value, -D key=value, --property key=value
// and --property=key=value. The key is split from the value on the first '='; later
// occurrences of a key win.
func parseGlobalProperties(args []string) map[string]string {
	props := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		var kv string
		switch {
		case arg == "-D" || arg == "--property":
			if i+1 < len(args) {
				i++
				kv = args[i]
			}
		case strings.HasPrefix(arg, "-D"):
			kv = arg[len("-D"):]
		case strings.HasPrefix(arg, "--property="):
			kv = arg[len("--property="):]
		default:
			continue
		}
		if key, value, ok := strings.Cut(kv, "="); ok && key != "" {
			props[key] = value
		}
	}
	return props
}

// globalPropertyArgSkip reports whether args[0] begins a -D/--property override and
// how many tokens it consumes: 1 when the value is attached (-Dkey=v, --property=k=v)
// and 2 when it follows as a separate token (-D k=v, --property k=v). It lets the
// command parser step over global property flags wherever they appear.
func globalPropertyArgSkip(args []string) (matched bool, skip int) {
	switch arg := args[0]; {
	case arg == "-D" || arg == "--property":
		return true, 2
	case strings.HasPrefix(arg, "-D") || strings.HasPrefix(arg, "--property="):
		return true, 1
	default:
		return false, 0
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
