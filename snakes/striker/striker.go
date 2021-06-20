package striker

import (
	"battlesnake/pkg/engine"
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

func HandleMove(w http.ResponseWriter, r *http.Request) {
	request := types.GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	var ml []types.SnakeMove
	for _, snake := range request.Board.Snakes {
		if snake.ID == request.You.ID {
			continue
		}
		ml = append(ml, bestMove(request.Board, snake, request.Turn))
	}

	response := bestMove(engine.MoveSnakes(request.Board, ml), request.You, request.Turn)

	fmt.Printf("MOVE: %s\n", response.Move)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func bestMove(b types.BoardState, you types.Snake, turn int) types.SnakeMove {
	g := grid.Make(b)
	snakes := grid.Snakes(b)

	for _, snake := range b.Snakes {
		if snake.ID == you.ID {
			continue
		}
		if len(snake.Body) < len(you.Body) {
			continue
		}

		h := snake.Body[0]
		for _, nghbr := range []types.Point{
			{Y: h.Y, X: h.X - 1},
			{Y: h.Y, X: h.X + 1},
			{Y: h.Y - 1, X: h.X},
			{Y: h.Y + 1, X: h.X},
		} {
			g.Cells[nghbr] = grid.Cell{Content: grid.ContentTypeAvoid}
		}
	}

	response := types.SnakeMove{ID: you.ID}
	amLongest := snakes[0].Name == you.Name
	if len(snakes) > 1 && amLongest && snakes[0].Length > snakes[1].Length {
		dir, dist := grid.FindPath(&g, you.Body[0], snakes[1].Head)
		if dist > 0 {
			fmt.Println(you.Name, "hunting")
			response.Move = dir
			return response
		}
	}

	if !amLongest || you.Health < 15 {
		dir, distance := shortestSafePathToFood(b, you, &g)
		if distance > 0 && (!amLongest || you.Health-distance <= 2) {
			fmt.Println(you.Name, "eating")
			response.Move = dir
			return response
		}
	}

	if turn > 1 {
		dir, dist := grid.FindPath(&g, you.Body[0], you.Body[len(you.Body)-1])
		if dist > 1 {
			fmt.Println(you.Name, "coiling")
			response.Move = dir
			return response
		}
	}

	fmt.Println(you.Name, "fleeing")
	h := you.Body[0]
	var biggestArea int
	for _, nghbr := range []struct {
		p   types.Point
		dir types.MoveDir
	}{
		{p: types.Point{Y: h.Y, X: h.X - 1}, dir: types.MoveDirLeft},
		{p: types.Point{Y: h.Y, X: h.X + 1}, dir: types.MoveDirRight},
		{p: types.Point{Y: h.Y - 1, X: h.X}, dir: types.MoveDirDown},
		{p: types.Point{Y: h.Y + 1, X: h.X}, dir: types.MoveDirUp},
	} {
		area := grid.FloodFill(&g, nghbr.p, len(you.Body))
		if area > biggestArea {
			biggestArea = area
			response.Move = nghbr.dir
		}
	}

	return response
}

func shortestSafePathToFood(b types.BoardState, you types.Snake, g *grid.Grid) (nextDir types.MoveDir, distance int) {
	sort.Slice(b.Food, func(i, j int) bool {
		return types.Distance(you.Body[0], b.Food[i]) < types.Distance(you.Body[0], b.Food[j])
	})

	for i := 0; i < len(b.Food); i++ {
		dir, dist := grid.FindPath(g, you.Body[0], b.Food[i])
		if dist == 0 {
			continue
		}

		_, dist2 := grid.FindPath(g, b.Food[i], you.Body[len(you.Body)-1])
		if dist2 == 0 {
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
