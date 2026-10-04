package server

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/coder/websocket"
)

// Client is a single WebSocket connection.
type Client struct {
	conn     *websocket.Conn
	hub      *Hub
	send     chan []byte
	room     *Room
	playerID string
}

func (c *Client) enqueue(m message) {
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		log.Printf("client %s: buffer pieno, messaggio scartato", c.playerID)
	}
}

func (c *Client) writePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-c.send:
			wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.conn.Write(wctx, websocket.MessageText, data)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump(ctx context.Context) {
	defer func() {
		if c.room != nil {
			c.room.leave(c)
		}
	}()
	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}
		var m message
		if err := json.Unmarshal(data, &m); err != nil {
			c.enqueue(message{Type: msgError, Message: "messaggio non valido"})
			continue
		}
		if m.Type == msgJoin {
			c.handleJoin(m)
			continue
		}
		if c.room == nil {
			c.enqueue(message{Type: msgError, Message: "invia prima un messaggio join"})
			continue
		}
		c.room.handle(c, m)
	}
}

func (c *Client) handleJoin(m message) {
	if c.room != nil {
		c.enqueue(message{Type: msgError, Message: "già connesso a una partita"})
		return
	}
	var room *Room
	if strings.TrimSpace(m.Game) == "" {
		room = c.hub.create()
	} else {
		room = c.hub.get(strings.ToUpper(strings.TrimSpace(m.Game)))
		if room == nil {
			c.enqueue(message{Type: msgError, Message: "partita non trovata"})
			return
		}
	}
	id, host, err := room.join(c, m.Name)
	if err != nil {
		c.enqueue(message{Type: msgError, Message: err.Error()})
		return
	}
	c.room = room
	c.playerID = id
	c.enqueue(message{Type: msgJoined, Game: room.code, PlayerID: id, Host: host})

	room.mu.Lock()
	room.broadcastLocked()
	room.mu.Unlock()
}
