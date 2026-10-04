package engine

import (
	"fmt"
	"sort"
	"strings"
)

// Word is a found word together with its score.
type Word struct {
	Text  string `json:"word"`
	Score int    `json:"score"`
}

// Solver finds every dictionary word that can be built from a rack, optionally
// crossing a fixed pattern. It is safe for concurrent use.
type Solver struct {
	dict     *Dictionary
	rackSize int
}

// New creates a Solver over an already loaded dictionary. rackSize is used
// only by callers (e.g. the game simulation); a value <= 0 defaults to 8.
func New(dict *Dictionary, rackSize int) *Solver {
	if rackSize <= 0 {
		rackSize = 8
	}
	return &Solver{dict: dict, rackSize: rackSize}
}

// RackSize returns the configured rack size.
func (s *Solver) RackSize() int { return s.rackSize }

// Dictionary exposes the underlying word list.
func (s *Solver) Dictionary() *Dictionary { return s.dict }

// FindWords returns every word that can be formed from caret, filtered by the
// optional crossing pattern. Results are sorted by score (descending) and then
// alphabetically, making the output deterministic.
func (s *Solver) FindWords(caret, crossing string) ([]Word, error) {
	rack, err := normalizeRack(caret)
	if err != nil {
		return nil, err
	}

	pattern, err := ParsePattern(crossing)
	if err != nil {
		return nil, err
	}

	fixed := pattern.FixedLetters()
	pool := rack + fixed

	var poolCounts [alphabetSize]int
	for i := 0; i < len(pool); i++ {
		poolCounts[pool[i]-'a']++
	}

	// A word can never be longer than the letter pool it is drawn from.
	maxLen := len(pool)
	if maxLen > s.dict.maxLen {
		maxLen = s.dict.maxLen
	}

	var found []Word
	for length := 1; length <= maxLen; length++ {
		for _, entry := range s.dict.byLength[length] {
			if !pattern.Matches(entry.text) {
				continue
			}
			if !fits(entry.counts, poolCounts) {
				continue
			}
			found = append(found, Word{
				Text:  entry.text,
				Score: s.Score(entry.text, len(fixed)),
			})
		}
	}

	sort.Slice(found, func(i, j int) bool {
		if found[i].Score != found[j].Score {
			return found[i].Score > found[j].Score
		}
		return found[i].Text < found[j].Text
	})
	return found, nil
}

// Score computes the value of a word. tilesFromRack is the number of letters
// taken from the rack, i.e. the word length minus the fixed crossing letters.
// The original length bonuses are preserved: 6 -> +10, 7 -> +30, 8 -> +50.
func (s *Solver) Score(word string, fixedCrossing int) int {
	total := 0
	for i := 0; i < len(word); i++ {
		total += letterValue(word[i])
	}

	switch len(word) - fixedCrossing {
	case 6:
		total += 10
	case 7:
		total += 30
	case 8:
		total += 50
	}
	return total
}

// fits reports whether every letter of the word is available in the pool.
func fits(word [alphabetSize]uint8, pool [alphabetSize]int) bool {
	for i := 0; i < alphabetSize; i++ {
		if int(word[i]) > pool[i] {
			return false
		}
	}
	return true
}

// normalizeRack lowercases a rack, ignores whitespace and underscores, and
// rejects anything that is not an a-z letter.
func normalizeRack(caret string) (string, error) {
	var b strings.Builder
	b.Grow(len(caret))
	for i := 0; i < len(caret); i++ {
		c := caret[i]
		switch {
		case c >= 'a' && c <= 'z':
			b.WriteByte(c)
		case c >= 'A' && c <= 'Z':
			b.WriteByte(c - 'A' + 'a')
		case c == ' ' || c == '\t' || c == '_':
			// ignored separators
		default:
			return "", fmt.Errorf("invalid rack character %q", c)
		}
	}
	return b.String(), nil
}
