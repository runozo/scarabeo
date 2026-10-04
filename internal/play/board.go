// Package play contains the authoritative Scarabeo game logic: board layout,
// tiles, move validation and scoring. It is used by the WebSocket server so
// that every rule is enforced server-side.
package play

// Board geometry and game constants.
const (
	Size          = 17
	CenterIndex   = 8
	RackSize      = 8
	MinPlayers    = 2
	MaxPlayers    = 4
	ScarabeoBonus = 100
)

// Jolly is the letter used for the wildcard tile ("scarabeo").
const Jolly = byte('?')

// CellType identifies the bonus of a board square.
type CellType uint8

const (
	Normal CellType = iota
	DoubleLetter
	TripleLetter
	DoubleWord
	TripleWord
	Center
)

// LetterMultiplier returns 2 for a double-letter square, 3 for a triple-letter
// square and 1 otherwise.
func (t CellType) LetterMultiplier() int {
	switch t {
	case DoubleLetter:
		return 2
	case TripleLetter:
		return 3
	default:
		return 1
	}
}

// WordMultiplier returns 2 for a double-word square, 3 for a triple-word square
// and 1 otherwise. The center square is a plain starting square.
func (t CellType) WordMultiplier() int {
	switch t {
	case DoubleWord:
		return 2
	case TripleWord:
		return 3
	default:
		return 1
	}
}

// Layout is the premium-square layout of the Scarabeo board, reconstructed from
// the original Editrice Giochi board.
var Layout [Size][Size]CellType

var layoutRows = [Size]string{
	"WNNNLNNNWNNNLNNNW",
	"NDNNNNTNNNTNNNNDN",
	"NNDNNNNLNLNNNNDNN",
	"NNNDNNNNLNNNNDNNN",
	"LNNNDNNNNNNNDNNNL",
	"NNNNNDNNNNNDNNNNN",
	"NTNNNNTNNNTNNNNTN",
	"NNLNNNNLNLNNNNLNN",
	"NNNLNNNNCNNNNLNNN",
	"NNLNNNNLNLNNNNLNN",
	"NTNNNNTNNNTNNNNTN",
	"NNNNNDNNNNNDNNNNN",
	"LNNNDNNNNNNNDNNNL",
	"NNNDNNNNLNNNNDNNN",
	"NNDNNNNLNLNNNNDNN",
	"NDNNNNTNNNTNNNNDN",
	"WNNNLNNNWNNNLNNNW",
}

func init() {
	for r, row := range layoutRows {
		if len(row) != Size {
			panic("play: invalid board row length")
		}
		for c := 0; c < Size; c++ {
			Layout[r][c] = cellType(row[c])
		}
	}
}

func cellType(code byte) CellType {
	switch code {
	case 'L':
		return DoubleLetter
	case 'T':
		return TripleLetter
	case 'D':
		return DoubleWord
	case 'W':
		return TripleWord
	case 'C':
		return Center
	default:
		return Normal
	}
}

// Values maps a letter to its Scarabeo point value.
var Values = map[byte]int{
	'A': 1, 'B': 4, 'C': 1, 'D': 4, 'E': 1, 'F': 4, 'G': 4, 'H': 8, 'I': 1,
	'L': 2, 'M': 2, 'N': 2, 'O': 1, 'P': 3, 'Q': 10, 'R': 1, 'S': 1, 'T': 1,
	'U': 4, 'V': 4, 'Z': 8, Jolly: 0,
}

// Counts maps a letter to the number of tiles in the bag.
var Counts = map[byte]int{
	'A': 12, 'B': 4, 'C': 7, 'D': 4, 'E': 12, 'F': 4, 'G': 4, 'H': 2, 'I': 12,
	'L': 6, 'M': 6, 'N': 6, 'O': 12, 'P': 4, 'Q': 2, 'R': 7, 'S': 7, 'T': 7,
	'U': 4, 'V': 4, 'Z': 2, Jolly: 2,
}

// Letters lists every tile letter (jolly last).
var Letters = []byte{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'Z', Jolly,
}

// LengthBonus returns the bonus for playing n tiles in one move.
func LengthBonus(n int) int {
	switch n {
	case 6:
		return 10
	case 7:
		return 30
	case 8:
		return 50
	default:
		return 0
	}
}
