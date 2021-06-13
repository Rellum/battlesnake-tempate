package main

import (
	"github.com/BattlesnakeOfficial/rules"
)

type grid map[rules.Point]rune

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

type BattlesnakeInfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
}

type GameRequest struct {
	Game  Game             `json:"game"`
	Turn  int              `json:"turn"`
	Board rules.BoardState `json:"board"`
	You   Battlesnake      `json:"you"`
}

type MoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout,omitempty"`
}
