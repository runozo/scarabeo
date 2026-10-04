package play

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
)

// Phase is the lifecycle stage of a game.
type Phase uint8

const (
	Lobby Phase = iota
	Playing
	Over
)

func (p Phase) String() string {
	switch p {
	case Lobby:
		return "lobby"
	case Playing:
		return "playing"
	default:
		return "over"
	}
}

// Tile is a letter tile on the board.
type Tile struct {
	Letter   byte
	Value    int
	Jolly    bool
	Assigned byte // letter represented by a jolly
}

// LetterOf returns the effective letter of a tile (the assigned letter for a
// jolly).
func (t *Tile) LetterOf() byte {
	if t.Jolly && t.Assigned != 0 {
		return t.Assigned
	}
	return t.Letter
}

// Cell is a board square with an optional tile.
type Cell struct {
	Type CellType
	Tile *Tile
}

// Player is a participant with a hidden rack and a score.
type Player struct {
	ID        string
	Name      string
	Rack      []byte
	Score     int
	Connected bool
}

// Placement is a single tile placed by a player in a move.
type Placement struct {
	Row      int    `json:"row"`
	Col      int    `json:"col"`
	Letter   string `json:"letter"`             // tile taken from the rack, "?" for jolly
	Assigned string `json:"assigned,omitempty"` // represented letter for a jolly
}

// MovePreview is the server's verdict on a set of placements, without applying
// them.
type MovePreview struct {
	Valid bool     `json:"valid"`
	Error string   `json:"error,omitempty"`
	Score int      `json:"score"`
	Words []string `json:"words"`
}

// MoveResult describes an accepted move.
type MoveResult struct {
	PlayerID      string   `json:"playerId"`
	Score         int      `json:"score"`
	Words         []string `json:"words"`
	Cells         [][2]int `json:"cells"`
	LengthBonus   int      `json:"lengthBonus"`
	ScarabeoBonus int      `json:"scarabeoBonus"`
}

// PlayerState is the public view of a player.
type PlayerState struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Score     int    `json:"score"`
	Tiles     int    `json:"tiles"`
	Connected bool   `json:"connected"`
	IsCurrent bool   `json:"isCurrent"`
}

// BoardTile is a placed tile in the public board view.
type BoardTile struct {
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	Letter string `json:"letter"`
	Jolly  bool   `json:"jolly,omitempty"`
}

// State is the per-player snapshot sent to a client.
type State struct {
	Game      string        `json:"game"`
	Phase     string        `json:"phase"`
	You       string        `json:"you"`
	Rack      []string      `json:"rack"`
	Board     []BoardTile   `json:"board"`
	Players   []PlayerState `json:"players"`
	Current   string        `json:"current"`
	Bag       int           `json:"bag"`
	Last      *MoveResult   `json:"last,omitempty"`
	Winner    string        `json:"winner,omitempty"`
	CanStart  bool          `json:"canStart"`
	MinStart  int           `json:"minStart"`
	MaxPlayer int           `json:"maxPlayers"`
}

// Game holds the complete authoritative state of a match.
type Game struct {
	Code    string
	board   [Size][Size]Cell
	players []*Player
	bag     []byte
	current int
	phase   Phase
	last    *MoveResult
	winner  string
	passes  int
	rng     *rand.Rand
}

// NewGame creates an empty game in the lobby phase.
func NewGame(code string, rng *rand.Rand) *Game {
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}
	g := &Game{Code: code, phase: Lobby, rng: rng}
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			g.board[r][c].Type = Layout[r][c]
		}
	}
	return g
}

// Phase returns the current phase.
func (g *Game) Phase() Phase { return g.phase }

// Winner returns the winning player ID when the game is over.
func (g *Game) Winner() string { return g.winner }

// Players returns the player list.
func (g *Game) Players() []*Player { return g.players }

// PlayerByID finds a player by ID.
func (g *Game) PlayerByID(id string) *Player {
	for _, p := range g.players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// PlayerByName finds a player by name (case-insensitive).
func (g *Game) PlayerByName(name string) *Player {
	for _, p := range g.players {
		if strings.EqualFold(p.Name, name) {
			return p
		}
	}
	return nil
}

// CurrentPlayerID returns the ID of the player whose turn it is.
func (g *Game) CurrentPlayerID() string {
	if g.current < 0 || g.current >= len(g.players) {
		return ""
	}
	return g.players[g.current].ID
}

// AddPlayer adds a player to the lobby.
func (g *Game) AddPlayer(id, name string) (*Player, error) {
	if g.phase != Lobby {
		return nil, errors.New("la partita è già iniziata")
	}
	if len(g.players) >= MaxPlayers {
		return nil, errors.New("partita piena (massimo 4 giocatori)")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = fmt.Sprintf("Giocatore %d", len(g.players)+1)
	}
	if g.PlayerByName(name) != nil {
		return nil, errors.New("nome già in uso")
	}
	p := &Player{ID: id, Name: name, Connected: true}
	g.players = append(g.players, p)
	return p, nil
}

// SetConnected updates a player's connection flag.
func (g *Game) SetConnected(id string, connected bool) {
	if p := g.PlayerByID(id); p != nil {
		p.Connected = connected
	}
}

// Start deals the racks and begins play.
func (g *Game) Start() error {
	if g.phase != Lobby {
		return errors.New("la partita è già iniziata")
	}
	if len(g.players) < MinPlayers {
		return fmt.Errorf("servono almeno %d giocatori", MinPlayers)
	}
	g.bag = g.newBag()
	for _, p := range g.players {
		g.refill(p)
	}
	g.current = 0
	g.phase = Playing
	g.last = nil
	g.winner = ""
	g.passes = 0
	return nil
}

// Restart resets the game keeping the same players, back to the lobby.
func (g *Game) Restart() {
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			g.board[r][c].Tile = nil
		}
	}
	for _, p := range g.players {
		p.Rack = nil
		p.Score = 0
	}
	g.bag = nil
	g.current = 0
	g.phase = Lobby
	g.last = nil
	g.winner = ""
	g.passes = 0
}

func (g *Game) newBag() []byte {
	b := make([]byte, 0, 130)
	for _, l := range Letters {
		for i := 0; i < Counts[l]; i++ {
			b = append(b, l)
		}
	}
	g.rng.Shuffle(len(b), func(i, j int) { b[i], b[j] = b[j], b[i] })
	return b
}

func (g *Game) refill(p *Player) {
	for len(p.Rack) < RackSize && len(g.bag) > 0 {
		n := len(g.bag) - 1
		p.Rack = append(p.Rack, g.bag[n])
		g.bag = g.bag[:n]
	}
}

// ---------------------------------------------------------------------------
// Validation and scoring
// ---------------------------------------------------------------------------

var (
	errNotYourTurn = errors.New("non è il tuo turno")
	errNoMove      = errors.New("nessuna tessera posata")
)

type coord struct{ r, c int }

// Preview validates a move and returns its score without applying it.
func (g *Game) Preview(playerID string, placements []Placement) MovePreview {
	score, words, _, _, _, err := g.analyze(playerID, placements)
	if err != nil {
		return MovePreview{Valid: false, Error: err.Error()}
	}
	return MovePreview{Valid: true, Score: score, Words: words, Error: ""}
}

// Apply validates and commits a move.
func (g *Game) Apply(playerID string, placements []Placement) (*MoveResult, error) {
	score, words, coords, lb, sb, err := g.analyze(playerID, placements)
	if err != nil {
		return nil, err
	}
	p := g.PlayerByID(playerID)

	// remove the tiles from the rack
	for _, pl := range placements {
		letter := pl.Letter
		if letter == "" {
			letter = string(Jolly)
		}
		removeByte(&p.Rack, letter[0])
	}

	// place the tiles on the board
	for _, pl := range placements {
		t := &Tile{Letter: pl.Letter[0], Value: Values[pl.Letter[0]]}
		if pl.Letter[0] == Jolly {
			t.Jolly = true
			t.Assigned = pl.Assigned[0]
			t.Value = Values[t.Assigned]
		}
		g.board[pl.Row][pl.Col].Tile = t
	}

	p.Score += score
	g.refill(p)

	res := &MoveResult{
		PlayerID:      playerID,
		Score:         score,
		Words:         words,
		Cells:         coords,
		LengthBonus:   lb,
		ScarabeoBonus: sb,
	}
	g.last = res
	g.passes = 0
	g.advance()
	g.checkOver()
	return res, nil
}

// Pass skips the current player's turn.
func (g *Game) Pass(playerID string) error {
	if g.phase != Playing {
		return errors.New("la partita non è in corso")
	}
	if g.CurrentPlayerID() != playerID {
		return errNotYourTurn
	}
	g.passes++
	g.advance()
	if g.passes >= len(g.players) {
		g.checkOver()
	}
	return nil
}

func (g *Game) advance() {
	if len(g.players) == 0 {
		return
	}
	g.current = (g.current + 1) % len(g.players)
}

func (g *Game) checkOver() {
	if g.phase != Playing {
		return
	}
	allEmpty := true
	for _, p := range g.players {
		if len(p.Rack) > 0 {
			allEmpty = false
			break
		}
	}
	allPassed := g.passes >= len(g.players)
	if !allPassed && !(allEmpty && len(g.bag) == 0) {
		return
	}
	g.phase = Over
	best := g.players[0]
	for _, p := range g.players {
		if p.Score > best.Score {
			best = p
		}
	}
	g.winner = best.ID
}

// analyze performs every rule check and returns the resulting score. It never
// mutates the game.
func (g *Game) analyze(playerID string, placements []Placement) (score int, words []string, coords [][2]int, lb, sb int, err error) {
	if g.phase != Playing {
		return 0, nil, nil, 0, 0, errors.New("la partita non è in corso")
	}
	if g.CurrentPlayerID() != playerID {
		return 0, nil, nil, 0, 0, errNotYourTurn
	}
	if len(placements) == 0 {
		return 0, nil, nil, 0, 0, errNoMove
	}
	p := g.PlayerByID(playerID)
	if p == nil {
		return 0, nil, nil, 0, 0, errors.New("giocatore sconosciuto")
	}

	// validate the letters and the rack ownership
	rack := append([]byte(nil), p.Rack...)
	seen := make(map[[2]int]bool, len(placements))
	placed := make(map[[2]int]*Tile, len(placements))
	for _, pl := range placements {
		if pl.Row < 0 || pl.Row >= Size || pl.Col < 0 || pl.Col >= Size {
			return 0, nil, nil, 0, 0, fmt.Errorf("casella fuori dalla plancia (%d,%d)", pl.Row, pl.Col)
		}
		key := [2]int{pl.Row, pl.Col}
		if seen[key] {
			return 0, nil, nil, 0, 0, errors.New("casella ripetuta nella mossa")
		}
		seen[key] = true
		if g.board[pl.Row][pl.Col].Tile != nil {
			return 0, nil, nil, 0, 0, errors.New("casella già occupata")
		}
		letter, lerr := parseLetter(pl.Letter)
		if lerr != nil {
			return 0, nil, nil, 0, 0, lerr
		}
		if !removeByte(&rack, letter) {
			return 0, nil, nil, 0, 0, fmt.Errorf("non hai la tessera %q", pl.Letter)
		}
		t := &Tile{Letter: letter, Value: Values[letter]}
		if letter == Jolly {
			assigned, aerr := parseLetter(pl.Assigned)
			if aerr != nil || assigned == Jolly {
				return 0, nil, nil, 0, 0, errors.New("lo scarabeo deve indicare una lettera valida")
			}
			t.Jolly = true
			t.Assigned = assigned
			t.Value = Values[assigned]
		}
		placed[key] = t
	}

	// temporary board with the new tiles applied
	var tmp [Size][Size]Cell = g.board
	for key, t := range placed {
		tmp[key[0]][key[1]].Tile = t
	}
	tileAt := func(r, c int) *Tile {
		if r < 0 || c < 0 || r >= Size || c >= Size {
			return nil
		}
		return tmp[r][c].Tile
	}

	// straight line
	rows := map[int]bool{}
	cols := map[int]bool{}
	for _, pl := range placements {
		rows[pl.Row] = true
		cols[pl.Col] = true
	}
	if len(rows) > 1 && len(cols) > 1 {
		return 0, nil, nil, 0, 0, errors.New("le tessere devono essere tutte sulla stessa riga o colonna")
	}
	horizontal := len(rows) == 1

	// contiguity between the first and the last placed tile
	ordered := make([]Placement, len(placements))
	copy(ordered, placements)
	if horizontal {
		sortByCol(ordered)
	} else {
		sortByRow(ordered)
	}
	fixed := ordered[0].Row
	start := ordered[0].Col
	end := ordered[len(ordered)-1].Col
	if !horizontal {
		fixed = ordered[0].Col
		start = ordered[0].Row
		end = ordered[len(ordered)-1].Row
	}
	for i := start; i <= end; i++ {
		r, c := fixed, i
		if !horizontal {
			r, c = i, fixed
		}
		if tileAt(r, c) == nil {
			return 0, nil, nil, 0, 0, errors.New("le tessere devono essere adiacenti, senza spazi vuoti")
		}
	}

	// collect the words formed (main word + crossings)
	words = g.collectWords(ordered, tileAt)

	// a move must form at least one word of two or more letters
	hasWord := false
	for _, w := range words {
		if len([]rune(w)) >= 2 {
			hasWord = true
			break
		}
	}
	if !hasWord {
		return 0, nil, nil, 0, 0, errors.New("la mossa non forma nessuna parola (servono almeno 2 lettere)")
	}

	if g.boardIsEmpty() {
		if !seen[[2]int{CenterIndex, CenterIndex}] {
			return 0, nil, nil, 0, 0, errors.New("la prima parola deve coprire il centro")
		}
	} else {
		connected := false
		for _, pl := range placements {
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				rr, cc := pl.Row+d[0], pl.Col+d[1]
				if !seen[[2]int{rr, cc}] && tileAt(rr, cc) != nil {
					connected = true
				}
			}
		}
		if !connected {
			return 0, nil, nil, 0, 0, errors.New("la mossa deve collegarsi ad almeno una lettera già presente")
		}
	}

	// score
	score, lb, sb = g.scoreMove(ordered, tileAt, placed, words)
	coords = make([][2]int, 0, len(placements))
	for _, pl := range placements {
		coords = append(coords, [2]int{pl.Row, pl.Col})
	}
	return score, words, coords, lb, sb, nil
}

func (g *Game) collectWords(ordered []Placement, tileAt func(int, int) *Tile) []string {
	seenWord := map[string]bool{}
	var words []string
	for _, pl := range ordered {
		for _, dir := range []int{0, 1} { // 0 = horizontal, 1 = vertical
			cells := runCells(pl.Row, pl.Col, dir, tileAt)
			if len(cells) < 2 {
				continue
			}
			var b strings.Builder
			key := fmt.Sprintf("%d:", dir)
			for _, cr := range cells {
				t := tileAt(cr.r, cr.c)
				b.WriteByte(t.LetterOf())
				key += fmt.Sprintf("%d,%d;", cr.r, cr.c)
			}
			if !seenWord[key] {
				seenWord[key] = true
				words = append(words, b.String())
			}
		}
	}
	return words
}

func (g *Game) scoreMove(ordered []Placement, tileAt func(int, int) *Tile, placed map[[2]int]*Tile, words []string) (total, lengthBonus, scarabeoBonus int) {
	seenWord := map[string]bool{}
	for _, pl := range ordered {
		for _, dir := range []int{0, 1} {
			cells := runCells(pl.Row, pl.Col, dir, tileAt)
			if len(cells) < 2 {
				continue
			}
			key := fmt.Sprintf("%d:", dir)
			for _, cr := range cells {
				key += fmt.Sprintf("%d,%d;", cr.r, cr.c)
			}
			if seenWord[key] {
				continue
			}
			seenWord[key] = true
			total += g.scoreWord(cells, tileAt, placed)
		}
	}
	lengthBonus = LengthBonus(len(ordered))
	total += lengthBonus
	for _, w := range words {
		if w == "SCARABEO" {
			scarabeoBonus = ScarabeoBonus
			total += scarabeoBonus
			break
		}
	}
	return total, lengthBonus, scarabeoBonus
}

func (g *Game) scoreWord(cells []coord, tileAt func(int, int) *Tile, placed map[[2]int]*Tile) int {
	sum := 0
	mult := 1
	for _, cr := range cells {
		t := tileAt(cr.r, cr.c)
		if _, isNew := placed[[2]int{cr.r, cr.c}]; isNew {
			sum += t.Value * g.board[cr.r][cr.c].Type.LetterMultiplier()
			mult *= g.board[cr.r][cr.c].Type.WordMultiplier()
		} else {
			sum += t.Value
		}
	}
	return sum * mult
}

func (g *Game) boardIsEmpty() bool {
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			if g.board[r][c].Tile != nil {
				return false
			}
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// State snapshot
// ---------------------------------------------------------------------------

// StateFor builds the snapshot visible to a player (their own rack only).
func (g *Game) StateFor(playerID string) State {
	st := State{
		Game:      g.Code,
		Phase:     g.phase.String(),
		You:       playerID,
		Current:   g.CurrentPlayerID(),
		Bag:       len(g.bag),
		Last:      g.last,
		Winner:    g.winner,
		CanStart:  g.phase == Lobby && len(g.players) >= MinPlayers,
		MinStart:  MinPlayers,
		MaxPlayer: MaxPlayers,
	}
	for _, p := range g.players {
		st.Players = append(st.Players, PlayerState{
			ID:        p.ID,
			Name:      p.Name,
			Score:     p.Score,
			Tiles:     len(p.Rack),
			Connected: p.Connected,
			IsCurrent: p.ID == g.CurrentPlayerID(),
		})
		if p.ID == playerID {
			for _, l := range p.Rack {
				st.Rack = append(st.Rack, string(l))
			}
		}
	}
	if st.Rack == nil {
		st.Rack = []string{}
	}
	st.Board = []BoardTile{}
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			if t := g.board[r][c].Tile; t != nil {
				st.Board = append(st.Board, BoardTile{
					Row:    r,
					Col:    c,
					Letter: string(t.LetterOf()),
					Jolly:  t.Jolly,
				})
			}
		}
	}
	return st
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func runCells(r, c, dir int, tileAt func(int, int) *Tile) []coord {
	dr, dc := 0, 0
	if dir == 0 {
		dc = 1
	} else {
		dr = 1
	}
	sr, sc := r, c
	for tileAt(sr-dr, sc-dc) != nil {
		sr -= dr
		sc -= dc
	}
	var cells []coord
	for tileAt(sr, sc) != nil {
		cells = append(cells, coord{sr, sc})
		sr += dr
		sc += dc
	}
	return cells
}

func parseLetter(s string) (byte, error) {
	if s == "?" {
		return Jolly, nil
	}
	if len(s) != 1 {
		return 0, fmt.Errorf("tessera non valida %q", s)
	}
	ch := byte(strings.ToUpper(s)[0])
	if _, ok := Values[ch]; !ok {
		return 0, fmt.Errorf("tessera non valida %q", s)
	}
	return ch, nil
}

func removeByte(rack *[]byte, letter byte) bool {
	for i, b := range *rack {
		if b == letter {
			*rack = append((*rack)[:i], (*rack)[i+1:]...)
			return true
		}
	}
	return false
}

func sortByCol(p []Placement) {
	for i := 1; i < len(p); i++ {
		for j := i; j > 0 && p[j].Col < p[j-1].Col; j-- {
			p[j], p[j-1] = p[j-1], p[j]
		}
	}
}

func sortByRow(p []Placement) {
	for i := 1; i < len(p); i++ {
		for j := i; j > 0 && p[j].Row < p[j-1].Row; j-- {
			p[j], p[j-1] = p[j-1], p[j]
		}
	}
}
