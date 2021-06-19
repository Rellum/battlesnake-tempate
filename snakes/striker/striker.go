package striker

import (
	"battlesnake/pkg/grid"
	"battlesnake/pkg/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/beefsack/go-astar"
)

const snakeInfo = `{
  "apiversion": "1",
  "author": "dogzbody",
  "color" : "#eab676",
  "head" : "default",
  "tail" : "default",
  "version" : "0.0.1-beta"
}`

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

type pather struct {
	grid *grid.Grid
	p    types.Point
}

func (p *pather) PathNeighbors() []astar.Pather {
	var res []astar.Pather
	if p.p.X > 0 {
		res = append(res, &pather{grid: p.grid, p: types.Point{Y: p.p.Y, X: p.p.X - 1}})
	}

	if p.p.Y > 0 {
		res = append(res, &pather{grid: p.grid, p: types.Point{Y: p.p.Y - 1, X: p.p.X}})
	}

	if p.p.X < p.grid.Width-1 {
		res = append(res, &pather{grid: p.grid, p: types.Point{Y: p.p.Y, X: p.p.X + 1}})
	}

	if p.p.Y < p.grid.Height-1 {
		res = append(res, &pather{grid: p.grid, p: types.Point{Y: p.p.Y + 1, X: p.p.X}})
	}
	return res
}

func (p *pather) PathNeighborCost(astar.Pather) float64 {
	return 1
}

func (p *pather) PathEstimatedCost(toPather astar.Pather) float64 {
	return types.Distance(p.p, toPather.(*pather).p)
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

	var nearestFood types.Point
	nearestFoodDistance := float64(request.Board.Width + request.Board.Height) // max distance
	for _, p := range request.Board.Food {
		d := types.Distance(request.You.Head, p)
		if d < nearestFoodDistance {
			nearestFood = p
			nearestFoodDistance = d
		}
	}

	g := grid.Make(request.Board)
	path, _, ok := astar.Path(&pather{grid: &g, p: request.You.Head}, &pather{grid: &g, p: nearestFood})
	if !ok {
		log.Fatal("no path found")
	}
	nextCell := path[0].(*pather).p
	var response types.SnakeMove
	if nextCell.Y > request.You.Head.Y {
		response.Move = "up"
	} else if nextCell.Y < request.You.Head.Y {
		response.Move = "down"
	} else if nextCell.X > request.You.Head.X {
		response.Move = "right"
	} else {
		response.Move = "left"
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
