package engine

import (
	"strings"
	"testing"
)

// testDictionary is intentionally tiny and clean so expectations stay obvious.
const testDictionary = `casa
caso
cassa
amo
amaca
cane
cassaforte
`

func mustLoad(t *testing.T, words string) *Dictionary {
	t.Helper()
	d, err := loadDictionary(strings.NewReader(words), LoadOptions{})
	if err != nil {
		t.Fatalf("loadDictionary: %v", err)
	}
	return d
}

func TestScore(t *testing.T) {
	s := New(mustLoad(t, testDictionary), 8)

	tests := []struct {
		word   string
		fixed  int
		expect int
	}{
		{"casa", 0, 4},          // 1+1+1+1
		{"qua", 0, 15},          // q10+u4+a1
		{"aaaaaa", 0, 6 + 10},   // 6 letters bonus
		{"aaaaaaa", 0, 7 + 30},  // 7 letters bonus
		{"aaaaaaaa", 0, 8 + 50}, // 8 letters bonus
		{"casa", 1, 4},          // 4-1=3, no bonus
		{"cassaforte", 0, 13},   // 10 letters, no bonus
	}

	for _, tc := range tests {
		if got := s.Score(tc.word, tc.fixed); got != tc.expect {
			t.Errorf("Score(%q, %d) = %d, want %d", tc.word, tc.fixed, got, tc.expect)
		}
	}
}

func TestLetterValueUnknown(t *testing.T) {
	for _, ch := range []byte{'j', 'k', 'w', 'x', 'y', '-', ' '} {
		if got := letterValue(ch); got != 0 {
			t.Errorf("letterValue(%q) = %d, want 0", ch, got)
		}
	}
}

func TestParsePattern(t *testing.T) {
	if _, err := ParsePattern("a_b"); err != nil {
		t.Fatalf("a_b should be valid: %v", err)
	}
	if _, err := ParsePattern("a b"); err != nil {
		t.Fatalf("a b should be valid: %v", err)
	}
	if _, err := ParsePattern("a1b"); err == nil {
		t.Fatal("a1b should be rejected")
	}

	p, _ := ParsePattern("s_")
	if got := p.FixedLetters(); got != "s" {
		t.Errorf("FixedLetters = %q, want %q", got, "s")
	}
	if !p.Matches("casa") {
		t.Error("s_ should match casa")
	}
	if p.Matches("cane") {
		t.Error("s_ should not match cane")
	}

	empty, _ := ParsePattern("")
	if !empty.Empty() || !empty.Matches("anything") {
		t.Error("empty pattern must match everything")
	}
}

func TestFindWordsSubset(t *testing.T) {
	s := New(mustLoad(t, testDictionary), 8)

	got, err := s.FindWords("aacs", "")
	if err != nil {
		t.Fatalf("FindWords: %v", err)
	}
	want := []string{"casa"}
	if !sameWords(got, want) {
		t.Fatalf("FindWords(aacs) = %v, want %v", wordList(got), want)
	}
}

func TestFindWordsCrossing(t *testing.T) {
	s := New(mustLoad(t, testDictionary), 8)

	// "s" is a fixed crossing letter and is added to the pool.
	got, err := s.FindWords("aac", "s")
	if err != nil {
		t.Fatalf("FindWords: %v", err)
	}
	if !containsWord(got, "casa") {
		t.Fatalf("expected casa in %v", wordList(got))
	}
	// cassa needs two s but the pool only has one.
	if containsWord(got, "cassa") {
		t.Fatalf("cassa should not be found: %v", wordList(got))
	}
}

func TestFindWordsWildcardAndOrdering(t *testing.T) {
	s := New(mustLoad(t, testDictionary), 8)

	got, err := s.FindWords("aacss", "s_")
	if err != nil {
		t.Fatalf("FindWords: %v", err)
	}
	if !containsWord(got, "cassa") || !containsWord(got, "casa") {
		t.Fatalf("expected casa and cassa in %v", wordList(got))
	}
	// cassa (5 pts) must rank before casa (4 pts).
	if got[0].Text != "cassa" {
		t.Fatalf("expected cassa first, got %v", wordList(got))
	}
}

func TestFindWordsRejectsBadRack(t *testing.T) {
	s := New(mustLoad(t, testDictionary), 8)
	if _, err := s.FindWords("a1c", ""); err == nil {
		t.Fatal("expected error for invalid rack")
	}
}

func TestDictionaryCleaning(t *testing.T) {
	dirty := strings.Join([]string{
		"",
		"   ",
		"Casa",      // uppercased duplicate
		"casa",      // kept
		"casa",      // duplicate, skipped
		"ca sa",     // embedded space, rejected
		"ca\x00sa",  // NUL byte, rejected
		"123",       // digits, rejected
		"abat-jour", // hyphen, rejected without AllowPunct
	}, "\n")

	d := mustLoad(t, dirty)
	if d.Stats.Loaded != 1 {
		t.Fatalf("loaded = %d, want 1", d.Stats.Loaded)
	}
	if _, ok := d.byLength[4]; !ok {
		t.Fatal("expected a 4-letter word")
	}
}

func TestDictionaryAllowPunct(t *testing.T) {
	d, err := loadDictionary(strings.NewReader("abat-jour\n"), LoadOptions{AllowPunct: true})
	if err != nil {
		t.Fatalf("loadDictionary: %v", err)
	}
	if d.Stats.Loaded != 1 {
		t.Fatalf("loaded = %d, want 1", d.Stats.Loaded)
	}
}

func BenchmarkFindWords(b *testing.B) {
	d, err := LoadDictionary("../../dicts/italia-1a", LoadOptions{})
	if err != nil {
		b.Fatal(err)
	}
	s := New(d, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.FindWords("asmmquosd", "a"); err != nil {
			b.Fatal(err)
		}
	}
}

func sameWords(got []Word, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i].Text != want[i] {
			return false
		}
	}
	return true
}

func containsWord(got []Word, want string) bool {
	for _, w := range got {
		if w.Text == want {
			return true
		}
	}
	return false
}

func wordList(got []Word) []string {
	out := make([]string, len(got))
	for i, w := range got {
		out[i] = w.Text
	}
	return out
}
