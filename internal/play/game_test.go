package play

import (
	"fmt"
	"math/rand"
	"testing"
)

func newTestGame(t *testing.T, racks ...string) *Game {
	t.Helper()
	g := NewGame("TEST", rand.New(rand.NewSource(1)))
	for i := range racks {
		id := fmt.Sprintf("p%d", i+1)
		if _, err := g.AddPlayer(id, fmt.Sprintf("P%d", i+1)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for i, r := range racks {
		g.players[i].Rack = []byte(r)
	}
	g.bag = nil
	return g
}

func pl(row, col int, letter string) Placement {
	return Placement{Row: row, Col: col, Letter: letter}
}

func TestLayoutSanity(t *testing.T) {
	for r, row := range layoutRows {
		if len(row) != Size {
			t.Fatalf("row %d has %d cells", r, len(row))
		}
	}
	if Layout[CenterIndex][CenterIndex] != Center {
		t.Fatal("center square missing")
	}
	// symmetry: 180 degree rotation
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			if Layout[r][c] != Layout[Size-1-r][Size-1-c] {
				t.Fatalf("board not 180-symmetric at %d,%d", r, c)
			}
		}
	}
}

func TestFirstMoveOffCenterRejected(t *testing.T) {
	g := newTestGame(t, "CASA", "CASA")
	_, err := g.Apply("p1", []Placement{pl(0, 0, "C"), pl(0, 1, "A"), pl(0, 2, "S")})
	if err == nil {
		t.Fatal("expected the first move to be rejected")
	}
}

func TestFirstMoveOnCenter(t *testing.T) {
	g := newTestGame(t, "CASA", "CASA")
	res, err := g.Apply("p1", []Placement{pl(8, 7, "C"), pl(8, 8, "A"), pl(8, 9, "S")})
	if err != nil {
		t.Fatalf("valid first move rejected: %v", err)
	}
	if res.Score != 3 {
		t.Fatalf("score = %d, want 3", res.Score)
	}
	if g.CurrentPlayerID() != "p2" {
		t.Fatalf("turn did not advance, current = %s", g.CurrentPlayerID())
	}
}

func TestScarabeoBonus(t *testing.T) {
	g := newTestGame(t, "SCARABEO", "CASA")
	word := "SCARABEO"
	placements := make([]Placement, 0, 8)
	for i := 0; i < 8; i++ {
		placements = append(placements, pl(8, 4+i, string(word[i])))
	}
	preview := g.Preview("p1", placements)
	if !preview.Valid {
		t.Fatalf("SCARABEO rejected: %s", preview.Error)
	}
	if preview.Score != 161 {
		t.Fatalf("SCARABEO preview = %d, want 161", preview.Score)
	}
	res, err := g.Apply("p1", placements)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.ScarabeoBonus != 100 || res.LengthBonus != 50 {
		t.Fatalf("bonuses = %d/%d, want 100/50", res.ScarabeoBonus, res.LengthBonus)
	}
}

func TestDisconnectedMoveRejected(t *testing.T) {
	g := newTestGame(t, "CASACASA", "AB")
	if _, err := g.Apply("p1", []Placement{pl(8, 8, "C"), pl(8, 9, "A"), pl(8, 10, "S")}); err != nil {
		t.Fatalf("first move: %v", err)
	}
	if _, err := g.Apply("p2", []Placement{pl(0, 0, "A"), pl(0, 1, "B")}); err == nil {
		t.Fatal("expected disconnected move to be rejected")
	}
}

func TestNotInLineRejected(t *testing.T) {
	g := newTestGame(t, "CAS", "AB")
	if _, err := g.Apply("p1", []Placement{pl(8, 8, "C"), pl(9, 9, "A")}); err == nil {
		t.Fatal("expected non-linear move to be rejected")
	}
}

func TestGapRejected(t *testing.T) {
	g := newTestGame(t, "CAS", "AB")
	if _, err := g.Apply("p1", []Placement{pl(8, 7, "C"), pl(8, 9, "S")}); err == nil {
		t.Fatal("expected a move with a gap to be rejected")
	}
}

func TestTileNotInRackRejected(t *testing.T) {
	g := newTestGame(t, "CAS", "AB")
	if _, err := g.Apply("p1", []Placement{pl(8, 8, "C"), pl(8, 9, "A"), pl(8, 10, "B")}); err == nil {
		t.Fatal("expected a tile not in the rack to be rejected")
	}
}

func TestNotYourTurn(t *testing.T) {
	g := newTestGame(t, "CASACASA", "CASA")
	if _, err := g.Apply("p1", []Placement{pl(8, 8, "C"), pl(8, 9, "A"), pl(8, 10, "S")}); err != nil {
		t.Fatalf("first move: %v", err)
	}
	if _, err := g.Apply("p1", []Placement{pl(9, 8, "C"), pl(9, 9, "A")}); err == nil {
		t.Fatal("expected out-of-turn move to be rejected")
	}
}

func TestJolly(t *testing.T) {
	g := newTestGame(t, "C?SA", "CASA")
	res, err := g.Apply("p1", []Placement{
		pl(8, 7, "C"),
		{Row: 8, Col: 8, Letter: "?", Assigned: "A"},
		pl(8, 9, "S"),
		pl(8, 10, "A"),
	})
	if err != nil {
		t.Fatalf("jolly move rejected: %v", err)
	}
	if res.Score != 4 {
		t.Fatalf("jolly score = %d, want 4", res.Score)
	}
	if res.Words[0] != "CASA" {
		t.Fatalf("word = %v, want CASA", res.Words)
	}
}

func TestPreviewDoesNotMutate(t *testing.T) {
	g := newTestGame(t, "CASA", "CASA")
	g.Preview("p1", []Placement{pl(8, 7, "C"), pl(8, 8, "A"), pl(8, 9, "S")})
	if !g.boardIsEmpty() {
		t.Fatal("Preview mutated the board")
	}
	if len(g.players[0].Rack) != 4 {
		t.Fatal("Preview mutated the rack")
	}
}

func TestStateHidesOtherRacks(t *testing.T) {
	g := newTestGame(t, "CASA", "ABCD")
	st := g.StateFor("p1")
	if len(st.Rack) != 4 {
		t.Fatalf("own rack = %d, want 4", len(st.Rack))
	}
	for _, p := range st.Players {
		if p.ID == "p2" && p.Tiles != 4 {
			t.Fatalf("p2 tiles = %d, want 4", p.Tiles)
		}
	}
	if len(st.Board) != 0 {
		t.Fatal("board should be empty")
	}
}
