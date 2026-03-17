/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			curr[j] = ins
			if del < curr[j] {
				curr[j] = del
			}
			if sub < curr[j] {
				curr[j] = sub
			}
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// suggest returns the closest matching command or group name for the given
// input within the specified parent group. It returns "" if no reasonable
// match is found (distance must be at most half the input length, with a
// minimum threshold of 2).
func (t *implCliApplication) suggest(parentGroup, input string) string {
	var candidates []string

	for _, grp := range t.groups[parentGroup] {
		candidates = append(candidates, grp.Group())
	}
	if aliasMap, ok := t.groupAliases[parentGroup]; ok {
		for alias := range aliasMap {
			candidates = append(candidates, alias)
		}
	}
	for _, cmd := range t.commands[parentGroup] {
		candidates = append(candidates, cmd.Command())
	}
	if aliasMap, ok := t.cmdAliases[parentGroup]; ok {
		for alias := range aliasMap {
			candidates = append(candidates, alias)
		}
	}

	best := ""
	bestDist := -1
	for _, c := range candidates {
		d := levenshtein(input, c)
		if bestDist < 0 || d < bestDist {
			bestDist = d
			best = c
		}
	}

	// Only suggest if the distance is reasonable: at most half the input
	// length, capped at a minimum threshold of 2.
	maxDist := len(input) / 2
	if maxDist < 2 {
		maxDist = 2
	}
	if bestDist > maxDist {
		return ""
	}
	return best
}
