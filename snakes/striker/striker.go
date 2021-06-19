package striker

import (
	"battlesnake/pkg/grid"
	"battlesnake/pkg/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

	g := grid.Make(request.Board)
	graph := grid.MakeGraph(g)
	snakes := grid.Snakes(request.Board)

	var response types.SnakeMove
	if len(snakes) > 1 && snakes[0].Name == request.You.Name {
		fmt.Println("hunting")
		dir, _, ok := grid.FindPath(graph, request.You.Body[0], snakes[1].Head)
		if ok {
			response.Move = dir
		}
	}

	if response.Move == types.MoveDirUnknown && request.You.Health < 15 {
		dir, distance := shortestSafePathToFood(request, graph)
		if float64(request.You.Health)-distance <= 2 {
			fmt.Println("eating")
			response.Move = dir
		}
	}

	if response.Move == types.MoveDirUnknown && request.Turn > 3 {
		dir, _, ok := grid.FindPath(graph, request.You.Body[0], request.You.Body[len(request.You.Body)-1])
		if ok {
			fmt.Println("coiling")
			response.Move = dir
		}
	}

	if response.Move == types.MoveDirUnknown {
		h := request.You.Body[0]
		for _, nghbr := range []struct {
			p   types.Point
			dir types.MoveDir
		}{
			{p: types.Point{Y: h.Y, X: h.X - 1}, dir: types.MoveDirLeft},
			{p: types.Point{Y: h.Y, X: h.X + 1}, dir: types.MoveDirRight},
			{p: types.Point{Y: h.Y - 1, X: h.X}, dir: types.MoveDirDown},
			{p: types.Point{Y: h.Y + 1, X: h.X}, dir: types.MoveDirUp},
		} {
			if g.Cells[nghbr.p].Content != grid.ContentTypeEmpty && g.Cells[nghbr.p].Content != grid.ContentTypeFood {
				continue
			}

			fmt.Println("fleeing")
			response.Move = nghbr.dir
		}
	}

	fmt.Printf("MOVE: %s\n", response.Move)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func shortestSafePathToFood(req types.GameRequest, gph *grid.Graph) (nextDir types.MoveDir, distance float64) {
	sort.Slice(req.Board.Food, func(i, j int) bool {
		return types.Distance(req.You.Body[0], req.Board.Food[i]) < types.Distance(req.You.Body[0], req.Board.Food[j])
	})

	for i := 0; i < len(req.Board.Food); i++ {
		dir, dist, ok := grid.FindPath(gph, req.You.Body[0], req.Board.Food[i])
		if !ok {
			continue
		}

		_, _, ok = grid.FindPath(gph, req.Board.Food[i], req.You.Body[len(req.You.Body)-1])
		if !ok {
			continue
		}

		return dir, dist
	}

	return types.MoveDirUnknown, 0
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
