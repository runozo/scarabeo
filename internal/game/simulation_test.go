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
