package kaa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/BattlesnakeOfficial/rules"

	"battlesnake/pkg/mcts"
	"battlesnake/pkg/types"
)

const snakeInfo = `{
  "apiversion": "1",
  "author": "dogzbody",
  "color" : "#eab676",
  "head" : "viper",
  "tail" : "rattle",
  "version" : "0.0.2-beta"
}`

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(snakeInfo))
	if err != nil {
		log.Fatal(err)
	}
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	request := types.GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// Nothing to respond with here
	fmt.Print("START\n")
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
	Game  Game          `json:"game"`
	Turn  int32         `json:"turn"`
	Board BoardResponse `json:"board"`
	You   Battlesnake   `json:"you"`
}

type BoardResponse struct {
	Height  int32         `json:"height"`
	Width   int32         `json:"width"`
	Food    []rules.Point `json:"food"`
	Hazards []rules.Point `json:"hazards"`
	Snakes  []rules.Snake `json:"snakes"`
}

func HandleMove(w http.ResponseWriter, r *http.Request) {
	var request GameRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
	defer cancel()

	board := rules.BoardState{
		Height: request.Board.Height,
		Width:  request.Board.Width,
		Food:   request.Board.Food,
		Snakes: request.Board.Snakes,
	}

	move, err := mcts.Search(ctx, board, request.Board.Hazards, request.You.ID, request.Turn)
	if err != nil {
		log.Fatal(err)
	}
	response := rules.SnakeMove{
		ID:   request.You.ID,
		Move: move,
	}

	fmt.Printf("MOVE: %s\n", response.Move)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

// HandleEnd is called when a game your Battlesnake was playing has ended.
// It's purely for informational purposes, no response required.
func HandleEnd(w http.ResponseWriter, r *http.Request) {
	request := types.GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// Nothing to respond with here
	fmt.Print("END\n")
}
