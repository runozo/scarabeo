package server

import "github.com/runozo/scarabeo/internal/play"

// message is the envelope exchanged over the WebSocket, in both directions.
type message struct {
	Type string `json:"type"`

	// client -> server
	Game       string           `json:"game,omitempty"`
	Name       string           `json:"name,omitempty"`
	Placements []play.Placement `json:"placements,omitempty"`

	// server -> client
	PlayerID string            `json:"playerId,omitempty"`
	Host     bool              `json:"host,omitempty"`
	Message  string            `json:"message,omitempty"`
	State    *play.State       `json:"state,omitempty"`
	Preview  *play.MovePreview `json:"preview,omitempty"`
	Result   *play.MoveResult  `json:"result,omitempty"`
}

// Message types.
const (
	// client -> server
	msgJoin    = "join"
	msgStart   = "start"
	msgMove    = "move"
	msgPreview = "preview"
	msgPass    = "pass"
	msgRestart = "restart"

	// server -> client
	msgJoined   = "joined"
	msgState    = "state"
	msgError    = "error"
	msgPreviewS = "preview"
)
