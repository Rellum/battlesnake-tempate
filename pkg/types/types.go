package types

import "math"

type Game struct {
	ID      string `json:"id"`
	Timeout int32  `json:"timeout"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Snake struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Health          int             `json:"health"`
	Body            []Point         `json:"body"`
	Head            Point           `json:"head"`
	Length          int             `json:"length"`
	Shout           string          `json:"shout"`
	EliminatedCause EliminatedCause `json:"eliminated_cause"`
	EliminatedBy    string          `json:"eliminated_by"`
}

type EliminatedCause string

const (
	NotEliminated                   EliminatedCause = ""
	EliminatedByCollision           EliminatedCause = "snake-collision"
	EliminatedBySelfCollision       EliminatedCause = "snake-self-collision"
	EliminatedByOutOfHealth         EliminatedCause = "out-of-health"
	EliminatedByHeadToHeadCollision EliminatedCause = "head-collision"
	EliminatedByOutOfBounds         EliminatedCause = "wall-collision"
)

type BoardState struct {
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Food   []Point `json:"food"`
	Snakes []Snake `json:"snakes"`
}

type InfoResponse struct {
	APIVersion string `json:"apiversion"`
	Version    string `json:"version"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
}

type GameRequest struct {
	Game  Game       `json:"game"`
	Turn  int        `json:"turn"`
	Board BoardState `json:"board"`
	You   Snake      `json:"you"`
}

type SnakeMove struct {
	ID    string  `json:"id"`
	Move  MoveDir `json:"move"`
	Shout string  `json:"shout,omitempty"`
}

type MoveDir string

const (
	MoveDirUnknown MoveDir = ""
	MoveDirUp      MoveDir = "up"
	MoveDirDown    MoveDir = "down"
	MoveDirLeft    MoveDir = "left"
	MoveDirRight   MoveDir = "right"
)

func Distance(t, f Point) float64 {
	dX := t.X - f.X
	dY := t.Y - f.Y
	return math.Abs(float64(dX)) + math.Abs(float64(dY))
}
