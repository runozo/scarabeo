package engine

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// LoadOptions controls how a raw word list is cleaned while loading.
type LoadOptions struct {
	// AllowPunct keeps hyphen and apostrophe, which some word lists use for
	// compound and elided forms (e.g. "abat-jour"). When false only a-z
	// words are kept.
	AllowPunct bool
}

// LoadStats reports how many lines were read, kept and discarded.
type LoadStats struct {
	Total   int
	Loaded  int
	Skipped int
}

// wordEntry is a normalized word together with its precomputed letter
// histogram, so that FindWords never has to recount a word.
type wordEntry struct {
	text   string
	counts [alphabetSize]uint8
}

// Dictionary is an immutable, cleaned word list indexed by word length.
type Dictionary struct {
	byLength map[int][]wordEntry
	maxLen   int
	Stats    LoadStats
}

// LoadDictionary reads, normalizes and indexes a word list from path.
func LoadDictionary(path string, opts LoadOptions) (*Dictionary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadDictionary(f, opts)
}

// loadDictionary is the reader-based core of LoadDictionary, shared with
// tests.
func loadDictionary(r io.Reader, opts LoadOptions) (*Dictionary, error) {
	d := &Dictionary{byLength: make(map[int][]wordEntry)}
	seen := make(map[string]struct{})

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		d.Stats.Total++

		word, ok := sanitize(sc.Text(), opts)
		if !ok {
			d.Stats.Skipped++
			continue
		}
		if _, dup := seen[word]; dup {
			d.Stats.Skipped++
			continue
		}
		seen[word] = struct{}{}

		entry := wordEntry{text: word, counts: countLetters(word)}
		d.byLength[len(word)] = append(d.byLength[len(word)], entry)
		if len(word) > d.maxLen {
			d.maxLen = len(word)
		}
		d.Stats.Loaded++
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return d, nil
}

// sanitize lowercases and validates a raw dictionary line. A line is rejected
// as a whole when it contains anything but a-z (plus hyphen/apostrophe when
// allowed), so that malformed data is never silently merged into a new word.
func sanitize(raw string, opts LoadOptions) (string, bool) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return "", false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			continue
		}
		if opts.AllowPunct && (c == '-' || c == '\'') {
			continue
		}
		return "", false
	}
	return s, true
}

// MaxLen returns the length of the longest loaded word.
func (d *Dictionary) MaxLen() int { return d.maxLen }
