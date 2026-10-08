package game

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/runozo/scarabeo/internal/engine"
)

// DefaultGlobalCrate is the full set of Scarabeo tiles: the 21 Italian letters
// with their Scarabeo counts plus the two jolly tiles ('?'), for 130 tiles in
// total. It mirrors play.Counts.
const DefaultGlobalCrate = "aaaaaaaaaaaabbbbcccccccddddeeeeeeeeeeee" +
	"ffffgggghhiiiiiiiiiiiillllllmmmmmmnnnnnnoo" +
	"ooooooooooppppqqrrrrrrrssssssstttttttuuuuvvvvzz" +
	"??"

// Config controls a simulation run.
type Config struct {
	RackSize int
	Seed     int64
	// MaxTurns stops the simulation after this many turns; <= 0 means no limit.
	MaxTurns int
}

// Result summarises a finished simulation.
type Result struct {
	Turns   int
	Players []*Player
}

// Simulate plays a full game between the named players, writing a turn-by-turn
// log to out. globalCrate is the shared tile pool. The simulation stops when no
// player can make a move (or when MaxTurns is reached).
func Simulate(solver *engine.Solver, globalCrate string, names []string, cfg Config, out io.Writer) Result {
	if cfg.RackSize <= 0 {
		cfg.RackSize = 8
	}
	rng := rand.New(rand.NewSource(cfg.Seed))
	crate := []byte(globalCrate)

	players := make([]*Player, len(names))
	for i, name := range names {
		players[i] = NewPlayer(name, solver)
	}

	turn := 0
	for {
		if cfg.MaxTurns > 0 && turn >= cfg.MaxTurns {
			break
		}
		fmt.Fprintf(out, "Turn %d\n", turn)

		moved := false
		for _, p := range players {
			crate = p.PickFromCrate(crate, rng, cfg.RackSize)
			if word, ok := p.PlayOneMove(); ok {
				moved = true
				fmt.Fprintf(out, "%s plays %s (%d) | total %d | rack %q\n",
					p.Name, word.Text, word.Score, p.Score, string(p.Crate))
			} else {
				fmt.Fprintf(out, "%s passes | total %d | rack %q\n",
					p.Name, p.Score, string(p.Crate))
			}
		}
		fmt.Fprintf(out, "Global crate: %d tiles\n", len(crate))
		turn++

		if !moved {
			break
		}
	}

	return Result{Turns: turn, Players: players}
}
