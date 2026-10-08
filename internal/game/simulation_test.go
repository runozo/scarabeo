package game

import (
	"strings"
	"testing"

	"github.com/runozo/scarabeo/internal/engine"
)

func testSolver(t *testing.T) *engine.Solver {
	t.Helper()
	d, err := engine.LoadDictionary("../../dicts/italia-1a", engine.LoadOptions{})
	if err != nil {
		t.Fatalf("load dictionary: %v", err)
	}
	return engine.New(d, 8)
}

func TestSimulateTerminatesAndScores(t *testing.T) {
	solver := testSolver(t)

	var out strings.Builder
	res := Simulate(solver, DefaultGlobalCrate, []string{"King Kong", "Hulk"},
		Config{Seed: 42, MaxTurns: 200}, &out)

	if res.Turns == 0 {
		t.Fatal("simulation made no turns")
	}
	for _, p := range res.Players {
		if p.Score <= 0 {
			t.Errorf("%s finished with score %d, expected > 0", p.Name, p.Score)
		}
	}
	if !strings.Contains(out.String(), "Global crate:") {
		t.Error("expected a turn log in the output")
	}
}

func TestDefaultGlobalCrateIsFullSet(t *testing.T) {
	if len(DefaultGlobalCrate) != 130 {
		t.Fatalf("DefaultGlobalCrate has %d tiles, want 130", len(DefaultGlobalCrate))
	}
	counts := map[rune]int{}
	for _, r := range DefaultGlobalCrate {
		counts[r]++
	}
	if counts['?'] != 2 {
		t.Fatalf("jolly tiles = %d, want 2", counts['?'])
	}
	if counts['i'] != 12 {
		t.Fatalf("i tiles = %d, want 12", counts['i'])
	}
}

func TestRemoveLetterJollyFallback(t *testing.T) {
	p := &Player{Crate: []byte("a?")}
	if !p.removeLetter('a') {
		t.Fatal("expected to remove the first a")
	}
	if p.removeLetter('a') {
		t.Fatal("a second a should not be present")
	}
	if !p.removeLetter(Jolly) {
		t.Fatal("expected to fall back to the jolly")
	}
	if len(p.Crate) != 0 {
		t.Fatalf("crate = %q, want empty", p.Crate)
	}
}

func TestPlayOneMoveConsumesTiles(t *testing.T) {
	solver := testSolver(t)
	p := NewPlayer("P", solver)
	p.Crate = []byte("cas?")
	word, ok := p.PlayOneMove()
	if !ok {
		t.Fatal("expected a move from the rack")
	}
	if got := len(p.Crate); got != 4-len(word.Text) {
		t.Fatalf("crate %q after playing %q has %d tiles, want %d", p.Crate, word.Text, got, 4-len(word.Text))
	}
	// the only words available here need the jolly, so it must have been spent
	for _, c := range p.Crate {
		if c == Jolly {
			t.Fatalf("jolly should have been spent on %q, crate = %q", word.Text, p.Crate)
		}
	}
}

func TestSimulateWithJollies(t *testing.T) {
	solver := testSolver(t)
	var out strings.Builder
	// a crate made only of jollies: the solver must still be able to play
	// words (each jolly stands for a real letter).
	res := Simulate(solver, "????????", []string{"A", "B"}, Config{Seed: 1, MaxTurns: 20}, &out)
	if res.Turns == 0 {
		t.Fatal("expected at least one turn")
	}
	if !strings.Contains(out.String(), "plays") {
		t.Fatal("expected at least one played word")
	}
}

func TestSimulateDeterministicWithSeed(t *testing.T) {
	solver := testSolver(t)

	var a, b strings.Builder
	ra := Simulate(solver, DefaultGlobalCrate, []string{"A", "B"}, Config{Seed: 7, MaxTurns: 20}, &a)
	rb := Simulate(solver, DefaultGlobalCrate, []string{"A", "B"}, Config{Seed: 7, MaxTurns: 20}, &b)

	if a.String() != b.String() {
		t.Fatal("same seed must produce the same game")
	}
	if ra.Players[0].Score != rb.Players[0].Score {
		t.Fatal("same seed must produce the same scores")
	}
}
