package game

import (
	"math/rand"

	"github.com/runozo/scarabeo/internal/engine"
)

// Player is a simulated Scarabeo player holding a rack, the words played so
// far and a running score.
type Player struct {
	Name  string
	Crate []byte
	Words []engine.Word
	Score int

	solver *engine.Solver
}

// NewPlayer creates an empty player backed by the given solver.
func NewPlayer(name string, solver *engine.Solver) *Player {
	return &Player{Name: name, solver: solver}
}

// PickFromCrate fills the rack up to rackSize with random letters drawn from
// gameCrate. It returns the remaining gameCrate.
func (p *Player) PickFromCrate(gameCrate []byte, rng *rand.Rand, rackSize int) []byte {
	for len(p.Crate) < rackSize && len(gameCrate) > 0 {
		i := rng.Intn(len(gameCrate))
		p.Crate = append(p.Crate, gameCrate[i])
		gameCrate = append(gameCrate[:i], gameCrate[i+1:]...)
	}
	return gameCrate
}

// PlayOneMove plays the highest scoring word available and removes its letters
// from the rack. It returns false when no word can be formed.
func (p *Player) PlayOneMove() (engine.Word, bool) {
	words, err := p.solver.FindWords(string(p.Crate), "")
	if err != nil || len(words) == 0 {
		return engine.Word{}, false
	}

	best := words[0]
	for i := 0; i < len(best.Text); i++ {
		p.removeLetter(best.Text[i])
	}
	p.Words = append(p.Words, best)
	p.Score += best.Score
	return best, true
}

func (p *Player) removeLetter(c byte) {
	for i, r := range p.Crate {
		if r == c {
			p.Crate = append(p.Crate[:i], p.Crate[i+1:]...)
			return
		}
	}
}
