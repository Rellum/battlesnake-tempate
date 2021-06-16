package main

import (
	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/cli/commands"
)

type grid map[rules.Point]cell

type cell struct {
	content rune
	ttl     int
}

type Game struct {
	ID      string `json:"id"`
	Timeout int32  `json:"timeout"`
}

type Battlesnake struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Health int32         `json:"health"`
	Body   []rules.Point `json:"body"`
	Head   rules.Point   `json:"head"`
	Length int32         `json:"length"`
	Shout  string        `json:"shout"`
}

type GameRequest struct {
	Game  Game                   `json:"game"`
	Turn  int32                  `json:"turn"`
	Board commands.BoardResponse `json:"board"`
	You   Battlesnake            `json:"you"`
}

type MoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout,omitempty"`
}
