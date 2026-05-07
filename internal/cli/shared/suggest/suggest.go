// Package suggest produces near-match command suggestions for typos.
package suggest

import (
	"sort"
	"strings"
)

type candidate struct {
	name  string
	score int
	dist  int
}

// Commands returns up to a few likely command suggestions for the given input.
// It is intentionally conservative: when no candidate is reasonably close, it
// returns nil.
func Commands(input string, candidates []string) []string {
	in := strings.ToLower(strings.TrimSpace(input))
	if in == "" {
		return nil
	}

	collected := make([]candidate, 0, len(candidates))
	for _, raw := range candidates {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" || name == in {
			continue
		}

		score := 0
		if strings.HasPrefix(name, in) || strings.HasPrefix(in, name) {
			score += 4
		}
		if strings.Contains(name, in) || strings.Contains(in, name) {
			score += 2
		}

		dist := levenshtein(in, name)
		switch {
		case dist == 0:
			score += 8
		case dist == 1:
			score += 3
		case dist == 2:
			score += 1
		}
		if score == 0 {
			continue
		}
		collected = append(collected, candidate{name: raw, score: score, dist: dist})
	}

	if len(collected) == 0 {
		return nil
	}

	sort.SliceStable(collected, func(i, j int) bool {
		if collected[i].score != collected[j].score {
			return collected[i].score > collected[j].score
		}
		if collected[i].dist != collected[j].dist {
			return collected[i].dist < collected[j].dist
		}
		return collected[i].name < collected[j].name
	})

	limit := 3
	if len(collected) < limit {
		limit = len(collected)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, collected[i].name)
	}
	return out
}

func levenshtein(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
