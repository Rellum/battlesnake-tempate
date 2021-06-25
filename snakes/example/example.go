package example

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"battlesnake/pkg/types"
)

const snakeInfo = `{
  "apiversion": "1",
  "author": "",
  "color" : "#888888",
  "head" : "default",
  "tail" : "default",
  "version" : "0.0.1-beta"
}` // TODO: Your Battlesnake username and attributes.

// HandleIndex is called when your Battlesnake is created and refreshed
// by play.battlesnake.com. BattlesnakeInfoResponse contains information about
// your Battlesnake, including what it should look like on the game board.
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(snakeInfo))
	if err != nil {
		log.Fatal(err)
	}
}

// HandleStart is called at the start of each game your Battlesnake is playing.
// The GameRequest object contains information about the game that's about to start.
// TODO: Use this function to decide how your Battlesnake is going to look on the board.
func HandleStart(w http.ResponseWriter, r *http.Request) {
	request := types.GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// Nothing to respond with here
	fmt.Print("START\n")
}

// HandleMove is called for each turn of each game.
// Valid responses are "up", "down", "left", or "right".
// TODO: Use the information in the GameRequest object to determine your next move.
func HandleMove(w http.ResponseWriter, r *http.Request) {
	request := types.GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// Choose a random direction to move in
	possibleMoves := []types.MoveDir{"up", "down", "left", "right"}
	move := possibleMoves[rand.Intn(len(possibleMoves))]

	response := types.SnakeMove{
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
