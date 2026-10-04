package server

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/runozo/scarabeo/internal/play"
)

// Room owns one match and its connected clients.
type Room struct {
	code string
	hub  *Hub

	mu      sync.Mutex
	game    *play.Game
	clients map[string]*Client
	hostID  string
	seq     int
}

func newRoom(code string, hub *Hub) *Room {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &Room{
		code:    code,
		hub:     hub,
		game:    play.NewGame(code, rng),
		clients: make(map[string]*Client),
	}
}

// join attaches a client to the room, creating a player in the lobby or
// reattaching to a disconnected player during a match.
func (r *Room) join(c *Client, name string) (playerID string, host bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name = strings.TrimSpace(name)

	if r.game.Phase() == play.Lobby {
		id := fmt.Sprintf("p%d", r.seq+1)
		r.seq++
		if _, err := r.game.AddPlayer(id, name); err != nil {
			return "", false, err
		}
		r.clients[id] = c
		if r.hostID == "" {
			r.hostID = id
		}
		return id, id == r.hostID, nil
	}

	// match in progress: allow reconnecting to a disconnected seat
	p := r.game.PlayerByName(name)
	if p == nil || p.Connected {
		return "", false, errors.New("partita già iniziata")
	}
	p.Connected = true
	r.clients[p.ID] = c
	return p.ID, p.ID == r.hostID, nil
}

func (r *Room) leave(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.playerID == "" {
		return
	}
	if _, ok := r.clients[c.playerID]; !ok {
		return
	}
	delete(r.clients, c.playerID)
	r.game.SetConnected(c.playerID, false)
	if len(r.clients) == 0 {
		r.hub.remove(r.code)
		return
	}
	r.broadcastLocked()
}

// handle dispatches a game command.
func (r *Room) handle(c *Client, m message) {
	switch m.Type {
	case msgStart:
		r.start(c)
	case msgPreview:
		r.preview(c, m)
	case msgMove:
		r.move(c, m)
	case msgPass:
		r.pass(c)
	case msgRestart:
		r.restart(c)
	default:
		c.enqueue(message{Type: msgError, Message: "comando sconosciuto: " + m.Type})
	}
}

func (r *Room) start(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.playerID != r.hostID {
		c.enqueue(message{Type: msgError, Message: "solo l'host può avviare la partita"})
		return
	}
	if err := r.game.Start(); err != nil {
		c.enqueue(message{Type: msgError, Message: err.Error()})
		return
	}
	r.broadcastLocked()
}

func (r *Room) restart(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.playerID != r.hostID {
		c.enqueue(message{Type: msgError, Message: "solo l'host può ricominciare"})
		return
	}
	r.game.Restart()
	r.broadcastLocked()
}

func (r *Room) preview(c *Client, m message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pv := r.game.Preview(c.playerID, m.Placements)
	c.enqueue(message{Type: msgPreviewS, Preview: &pv})
}

func (r *Room) move(c *Client, m message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.game.Apply(c.playerID, m.Placements); err != nil {
		c.enqueue(message{Type: msgError, Message: err.Error()})
		return
	}
	r.broadcastLocked()
}

func (r *Room) pass(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.game.Pass(c.playerID); err != nil {
		c.enqueue(message{Type: msgError, Message: err.Error()})
		return
	}
	r.broadcastLocked()
}

func (r *Room) broadcastLocked() {
	for id, c := range r.clients {
		st := r.game.StateFor(id)
		c.enqueue(message{Type: msgState, State: &st})
	}
}
