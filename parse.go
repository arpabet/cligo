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
