package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/runozo/scarabeo/internal/play"
)

type testClient struct {
	t    *testing.T
	conn *websocket.Conn
}

func dial(t *testing.T, url string) *testClient {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return &testClient{t: t, conn: conn}
}

func (c *testClient) send(m message) {
	c.t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		c.t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.conn.Write(ctx, websocket.MessageText, data); err != nil {
		c.t.Fatalf("write: %v", err)
	}
}

func (c *testClient) recv() message {
	c.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, data, err := c.conn.Read(ctx)
	if err != nil {
		c.t.Fatalf("read: %v", err)
	}
	var m message
	if err := json.Unmarshal(data, &m); err != nil {
		c.t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func (c *testClient) waitFor(typ string) message {
	c.t.Helper()
	for i := 0; i < 30; i++ {
		m := c.recv()
		if m.Type == typ {
			return m
		}
	}
	c.t.Fatalf("did not receive message %q", typ)
	return message{}
}

func (c *testClient) waitPhase(phase string) *play.State {
	c.t.Helper()
	for i := 0; i < 30; i++ {
		m := c.recv()
		if m.Type == msgState && m.State != nil && m.State.Phase == phase {
			return m.State
		}
	}
	c.t.Fatalf("did not receive state phase %q", phase)
	return nil
}

// rejectAll rejects every word, so tests can assert that word checking is
// enforced server-side.
type rejectAll struct{}

func (rejectAll) Has(string) bool { return false }

func newTestServer(t *testing.T, validator play.WordValidator) (*httptest.Server, string) {
	t.Helper()
	srv := httptest.NewServer(New("../../web", validator).Handler())
	t.Cleanup(srv.Close)
	return srv, "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
}

func TestStaticAssetsAreNotCached(t *testing.T) {
	srv, _ := newTestServer(t, nil)

	resp, err := http.Get(srv.URL + "/app.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("Cache-Control = %q, want no-cache", got)
	}
}

func TestLobbyAndStart(t *testing.T) {
	_, url := newTestServer(t, nil)

	a := dial(t, url)
	a.send(message{Type: msgJoin, Name: "Alice"})
	joined := a.waitFor(msgJoined)
	if joined.Game == "" {
		t.Fatal("expected a game code")
	}
	if !joined.Host {
		t.Fatal("first player should be host")
	}
	a.waitPhase("lobby")

	b := dial(t, url)
	b.send(message{Type: msgJoin, Game: joined.Game, Name: "Bob"})
	bj := b.waitFor(msgJoined)
	if bj.Host {
		t.Fatal("second player should not be host")
	}

	a.send(message{Type: msgStart})
	sa := a.waitPhase("playing")
	if len(sa.Rack) != 8 {
		t.Fatalf("rack = %d, want 8", len(sa.Rack))
	}
	if len(sa.Players) != 2 {
		t.Fatalf("players = %d, want 2", len(sa.Players))
	}
}

func TestMoveValidationIsServerSide(t *testing.T) {
	_, url := newTestServer(t, nil)

	a := dial(t, url)
	a.send(message{Type: msgJoin, Name: "Alice"})
	joined := a.waitFor(msgJoined)
	a.waitPhase("lobby")

	b := dial(t, url)
	b.send(message{Type: msgJoin, Game: joined.Game, Name: "Bob"})
	b.waitFor(msgJoined)

	a.send(message{Type: msgStart})
	sa := a.waitPhase("playing")
	if len(sa.Rack) == 0 {
		t.Fatal("empty rack")
	}

	// off-center first move using a real tile from the rack
	a.send(message{Type: msgMove, Placements: []play.Placement{{Row: 0, Col: 0, Letter: sa.Rack[0]}}})
	errMsg := a.waitFor(msgError)
	if errMsg.Message == "" {
		t.Fatal("expected the server to reject the off-center first move")
	}

	// out-of-turn move from Bob
	b.waitPhase("playing")
	b.send(message{Type: msgMove, Placements: []play.Placement{{Row: 8, Col: 8, Letter: "A"}}})
	errB := b.waitFor(msgError)
	if errB.Message == "" {
		t.Fatal("expected the server to reject the out-of-turn move")
	}
}

func TestPreviewIsServerSide(t *testing.T) {
	_, url := newTestServer(t, nil)

	a := dial(t, url)
	a.send(message{Type: msgJoin, Name: "Alice"})
	joined := a.waitFor(msgJoined)
	a.waitPhase("lobby")

	b := dial(t, url)
	b.send(message{Type: msgJoin, Game: joined.Game, Name: "Bob"})
	b.waitFor(msgJoined)

	a.send(message{Type: msgStart})
	a.waitPhase("playing")

	a.send(message{Type: msgPreview, Placements: []play.Placement{{Row: 0, Col: 0, Letter: "A"}}})
	pv := a.waitFor(msgPreviewS)
	if pv.Preview == nil {
		t.Fatal("missing preview")
	}
	if pv.Preview.Valid {
		t.Fatal("off-center preview should be invalid")
	}
}

func TestWordValidationIsServerSide(t *testing.T) {
	_, url := newTestServer(t, rejectAll{})

	a := dial(t, url)
	a.send(message{Type: msgJoin, Name: "Alice"})
	joined := a.waitFor(msgJoined)
	a.waitPhase("lobby")

	b := dial(t, url)
	b.send(message{Type: msgJoin, Game: joined.Game, Name: "Bob"})
	b.waitFor(msgJoined)

	a.send(message{Type: msgStart})
	sa := a.waitPhase("playing")

	// pick two non-jolly letters from the rack and play them across the center
	letters := make([]string, 0, 2)
	for _, l := range sa.Rack {
		if l != "?" {
			letters = append(letters, l)
		}
		if len(letters) == 2 {
			break
		}
	}
	if len(letters) < 2 {
		t.Fatal("not enough letters on the rack")
	}

	a.send(message{Type: msgMove, Placements: []play.Placement{
		{Row: 8, Col: 8, Letter: letters[0]},
		{Row: 8, Col: 9, Letter: letters[1]},
	}})
	errMsg := a.waitFor(msgError)
	if !strings.Contains(errMsg.Message, "parola non valida") {
		t.Fatalf("expected a word validation error, got %q", errMsg.Message)
	}
}
