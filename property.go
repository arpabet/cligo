/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import "go.arpabet.com/glue"

// cliPropertyResolverPriority places command-line overrides at the top of the
// property-resolution stack — above glue's dotenv (300), environment (200) and
// file/map (100) resolvers — so a -D/--property value wins over every other
// source. This realises the "flags > env vars > config file > defaults" order.
const cliPropertyResolverPriority = 1000

// cliPropertyResolver is a glue.PropertyResolver backed by the key=value pairs
// passed on the command line via -D/--property. It is registered automatically by
// Run when any such flags are present.
type cliPropertyResolver struct {
	props map[string]string
}

// compile-time checks: the resolver must satisfy glue's resolver interfaces.
var (
	_ glue.PropertyResolver           = (*cliPropertyResolver)(nil)
	_ glue.EnumerablePropertyResolver = (*cliPropertyResolver)(nil)
)

func (r *cliPropertyResolver) Priority() int { return cliPropertyResolverPriority }

func (r *cliPropertyResolver) GetProperty(key string) (string, bool) {
	v, ok := r.props[key]
	return v, ok
}

// Keys implements glue.EnumerablePropertyResolver so command-line overrides also
// participate in prefix map injection (value:"prefix=X").
func (r *cliPropertyResolver) Keys() []string {
	keys := make([]string, 0, len(r.props))
	for k := range r.props {
		keys = append(keys, k)
	}
	return keys
}
