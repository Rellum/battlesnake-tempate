package silkworm

import (
	"battlesnake/pkg/grid"
	"battlesnake/pkg/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	response := bestMove(request.Board, request.You, request.Turn)
	if response.Move == types.MoveDirUnknown {
		response = safestMove(request.Board, request.You)
	}

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
			if !grid.IsValid(&g, nghbr) {
				continue
			}
			g.Cells[nghbr] = grid.Cell{Content: grid.ContentTypeAvoid}
		}
	}

	amShortest := snakes[len(snakes)-1].Name == you.Name
	if !amShortest && len(snakes) > 1 && snakes[len(snakes)-2].Length == snakes[len(snakes)-1].Length {
		amShortest = true
	}

	type dest struct {
		dir  types.MoveDir
		dist int
	}
	dests := make(map[string]dest)

	response := types.SnakeMove{ID: you.ID}
	if turn > 1 {
		dir, dist := grid.FindPath(&g, you.Body[0], you.Body[len(you.Body)-1])
		dests["tail"] = dest{dir: dir, dist: dist}
	}

	dir, dist := shortestSafePathToFood(b, you, &g)
	dests["food"] = dest{dir: dir, dist: dist}

	if len(snakes) > 1 {
		dir, dist := closestVictim(b, you, &g)
		dests["hunt"] = dest{dir: dir, dist: dist}
	}

	if amShortest || you.Health-dests["food"].dist <= 2 {
		fmt.Println(you.Name, "eating", dests["food"].dir)
		response.Move = dests["food"].dir
		return response
	}

	if dests["tail"].dist > 0 && dests["hunt"].dist > 0 && dests["tail"].dist+1 > dests["hunt"].dist {
		fmt.Println(you.Name, "hunting", dests["hunt"].dir)
		response.Move = dests["hunt"].dir
		return response
	}

	if dests["tail"].dist > 0 {
		fmt.Println(you.Name, "coiling", dests["tail"].dir)
		response.Move = dests["tail"].dir
		return response
	}

	return response
}

func safestMove(b types.BoardState, you types.Snake) types.SnakeMove {
	g := grid.Make(b)

	response := types.SnakeMove{ID: you.ID}

	fmt.Println(you.Name, "wondering")
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
	var bestDir types.MoveDir
	var bestDist = b.Width * b.Height
	for i := 0; i < len(b.Food); i++ {
		dir, dist := grid.FindPath(g, you.Body[0], b.Food[i])
		if dist == 0 {
			continue
		}
		if dist > bestDist {
			continue
		}

		_, dist2 := grid.FindPath(g, b.Food[i], you.Body[len(you.Body)-1])
		if dist2 == 0 {
			continue
		}

		bestDist = dist
		bestDir = dir
	}

	return bestDir, bestDist
}

func closestVictim(b types.BoardState, you types.Snake, g *grid.Grid) (nextDir types.MoveDir, distance int) {
	var bestDir types.MoveDir
	var bestDist = b.Width * b.Height
	for _, snake := range b.Snakes {
		if len(snake.Body) >= len(you.Body) {
			continue
		}

		dir, dist := grid.FindPath(g, you.Body[0], snake.Body[0])
		if dist == 0 {
			continue
		}
		if dist > bestDist {
			continue
		}

		bestDist = dist
		bestDir = dir
	}

	return bestDir, bestDist
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
