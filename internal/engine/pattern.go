package engine

import (
	"fmt"
	"strings"
)

// Pattern is a crossing pattern: a sequence of fixed letters and wildcards.
// A wildcard (underscore or space) matches any lowercase letter. A pattern
// matches a word when it occurs as a contiguous substring of the word.
//
// The empty pattern matches every word.
type Pattern struct {
	tokens []patternToken
	empty  bool
}

type patternToken struct {
	wildcard bool
	ch       byte
}

// ParsePattern validates and parses a crossing pattern such as "a_b" or
// "a b". Uppercase letters are folded to lowercase; underscore and space are
// wildcards. Any other character is rejected.
func ParsePattern(s string) (Pattern, error) {
	if s == "" {
		return Pattern{empty: true}, nil
	}

	tokens := make([]patternToken, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '_' || c == ' ':
			tokens = append(tokens, patternToken{wildcard: true})
		case c >= 'a' && c <= 'z':
			tokens = append(tokens, patternToken{ch: c})
		case c >= 'A' && c <= 'Z':
			tokens = append(tokens, patternToken{ch: c - 'A' + 'a'})
		default:
			return Pattern{}, fmt.Errorf("invalid crossing character %q", c)
		}
	}
	return Pattern{tokens: tokens}, nil
}

// Empty reports whether the pattern matches every word.
func (p Pattern) Empty() bool { return p.empty }

// FixedLetters returns the non-wildcard letters of the pattern, in order.
// These letters must be present on the board and are therefore added to the
// solver's letter pool.
func (p Pattern) FixedLetters() string {
	if p.empty {
		return ""
	}
	var b strings.Builder
	b.Grow(len(p.tokens))
	for _, t := range p.tokens {
		if !t.wildcard {
			b.WriteByte(t.ch)
		}
	}
	return b.String()
}

// Matches reports whether the pattern occurs as a contiguous substring of
// word. A wildcard only matches a lowercase a-z letter.
func (p Pattern) Matches(word string) bool {
	if p.empty {
		return true
	}
	n := len(p.tokens)
	if n == 0 || len(word) < n {
		return false
	}
	for start := 0; start+n <= len(word); start++ {
		if p.matchAt(word, start) {
			return true
		}
	}
	return false
}

func (p Pattern) matchAt(word string, start int) bool {
	for i, t := range p.tokens {
		c := word[start+i]
		if t.wildcard {
			if c < 'a' || c > 'z' {
				return false
			}
			continue
		}
		if c != t.ch {
			return false
		}
	}
	return true
}
