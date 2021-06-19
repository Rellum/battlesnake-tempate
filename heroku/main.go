package main

import (
	"battlesnake/snakes/striker"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/cli/commands"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var snakeInfo = `{
  "apiversion": "1",
  "author": "dogzbody",
  "color" : "#5499C7",
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
		log.Fatal().Err(err).Msg("")
	}
}

// HandleStart is called at the start of each game your Battlesnake is playing.
// The GameRequest object contains information about the game that's about to start.
// TODO: Use this function to decide how your Battlesnake is going to look on the board.
func HandleStart(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Nothing to respond with here
	fmt.Print("START\n")
}

// HandleMove is called for each turn of each game.
// Valid responses are "up", "down", "left", or "right".
// TODO: Use the information in the GameRequest object to determine your next move.
func HandleMove(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	log.Debug().Msg(string(b))

	request := GameRequest{}
	err = json.Unmarshal(b, &request)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	mr, err := doMove(request)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	response := MoveResponse{
		Move: mr.best(),
	}

	fmt.Printf("MOVE: %s\n", response.Move)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

type moveResult struct {
	u, d, l, r int
}

func (mr *moveResult) best() string {
	var res string
	var highest int

	for _, dir := range []string{"up", "down", "left", "right"} {
		score := mr.getScore(dir)
		if score <= highest {
			continue
		}

		highest = score
		res = dir
	}

	return res
}

func (mr *moveResult) getScore(dir string) int {
	switch dir {
	case "up":
		return mr.u
	case "down":
		return mr.d
	case "left":
		return mr.l
	case "right":
		return mr.r
	}
	return 0
}

func doMove(request GameRequest) (*moveResult, error) {
	var ruleset rules.Ruleset = &rules.StandardRuleset{
		FoodSpawnChance: 15,
		MinimumFood:     1,
	}
	depth := 2
	if len(request.Board.Snakes) == 1 {
		ruleset = &rules.SoloRuleset{rules.StandardRuleset{
			FoodSpawnChance: 15,
			MinimumFood:     1,
		}}
		depth = 3
	}
	if len(request.Board.Hazards) > 0 {
		ruleset = &rules.RoyaleRuleset{
			StandardRuleset: rules.StandardRuleset{
				FoodSpawnChance: 15,
				MinimumFood:     1,
			},
			Seed:              time.Now().UTC().UnixNano(),
			Turn:              request.Turn,
			ShrinkEveryNTurns: 10,
			DamagePerTurn:     1,
		}
	}

	res, err := simulate(ruleset, &rules.BoardState{
		Height: request.Board.Height,
		Width:  request.Board.Width,
		Food:   pointFromCoordSlice(request.Board.Food),
		Snakes: buildSnakes(request.Board.Snakes),
	}, request.You.ID, depth)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func simulate(rs rules.Ruleset, state *rules.BoardState, me string, depth int) (moveResult, error) {
	if depth == 0 {
		return moveResult{}, nil
	}
	grid := makeAnonymousGrid(*state)
	snakes := make(map[string]struct {
		len  int
		head rules.Point
	})
	for _, snake := range state.Snakes {
		snakes[snake.ID] = struct {
			len  int
			head rules.Point
		}{len: len(snake.Body), head: snake.Body[0]}
	}
	var mu sync.Mutex
	res := moveResult{
		u: math.MinInt64,
		d: math.MinInt64,
		l: math.MinInt64,
		r: math.MinInt64,
	}
	var eg errgroup.Group
	for _, ml := range allMoves(state.Snakes, nil) {
		ml := ml
		eg.Go(func() error {
			for _, move := range ml {
				if floodFill(grid, nextPos(snakes[move.ID].head, move.Move), nil, snakes[move.ID].len, 0) < snakes[move.ID].len {
					return nil
				}
			}

			newState, err := rs.CreateNextBoardState(state, ml)
			if err != nil {
				log.Fatal().Err(err).Stack().Msg("")
			}

			score, err := scoreTurn(state, newState, me)
			if err != nil {
				return err
			}

			mr, err := simulate(rs, newState, me, depth-1)
			if err != nil {
				return err
			}

			score += mr.u + mr.d + mr.r + mr.l
			var dir string
			for _, move := range ml {
				if move.ID == me {
					dir = move.Move
					break
				}
			}

			mu.Lock()
			defer mu.Unlock()
			switch dir {
			case "up":
				res.u += score
			case "down":
				res.d += score
			case "left":
				res.l += score
			case "right":
				res.r += score
			}
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return moveResult{}, err
	}
	mu.Lock()
	defer mu.Unlock()
	return res, nil
}

func coordFromPoint(pt rules.Point) commands.Coord {
	return commands.Coord{X: pt.X, Y: pt.Y}
}

func coordFromPointSlice(ptArray []rules.Point) []commands.Coord {
	a := make([]commands.Coord, 0)
	for _, pt := range ptArray {
		a = append(a, coordFromPoint(pt))
	}
	return a
}

func pointFromCoord(pt commands.Coord) rules.Point {
	return rules.Point{X: pt.X, Y: pt.Y}
}

func pointFromCoordSlice(ptArray []commands.Coord) []rules.Point {
	a := make([]rules.Point, 0)
	for _, pt := range ptArray {
		a = append(a, pointFromCoord(pt))
	}
	return a
}

func snakeFromSnakeResponse(snake commands.SnakeResponse) rules.Snake {
	return rules.Snake{
		ID:     snake.Id,
		Health: snake.Health,
		Body:   pointFromCoordSlice(snake.Body),
	}
}

func buildSnakes(snakes []commands.SnakeResponse) []rules.Snake {
	var a []rules.Snake
	for _, snake := range snakes {
		a = append(a, snakeFromSnakeResponse(snake))
	}
	return a
}

func snakeResponseFromSnake(snake rules.Snake) commands.SnakeResponse {
	return commands.SnakeResponse{
		Id:     snake.ID,
		Health: snake.Health,
		Body:   coordFromPointSlice(snake.Body),
		Head:   coordFromPoint(snake.Body[0]),
		Length: int32(len(snake.Body)),
	}
}

func buildSnakesResponse(snakes []rules.Snake) []commands.SnakeResponse {
	var a []commands.SnakeResponse
	for _, snake := range snakes {
		a = append(a, snakeResponseFromSnake(snake))
	}
	return a
}

func scoreTurn(p, t *rules.BoardState, me string) (int, error) {
	var res int

	var ps rules.Snake
	for _, snake := range p.Snakes {
		if snake.ID != me {
			continue
		}
		ps = snake
	}

	var longest int
	var ts rules.Snake
	for _, snake := range t.Snakes {
		if len(snake.Body) > longest {
			longest = len(snake.Body)
		}
		if snake.ID != me {
			continue
		}
		ts = snake
	}

	if ts.EliminatedCause != rules.NotEliminated {
		return lostGame, nil
	}

	var pStrikeDist float64
	for _, snake := range p.Snakes {
		if len(snake.Body) < len(ps.Body) {
			pStrikeDist += math.Abs(float64(snake.Body[0].X - ps.Body[0].X))
			pStrikeDist += math.Abs(float64(snake.Body[0].Y - ps.Body[0].Y))
		} else {
			pStrikeDist -= math.Abs(float64(snake.Body[0].X - ps.Body[0].X))
			pStrikeDist -= math.Abs(float64(snake.Body[0].Y - ps.Body[0].Y))
		}
	}

	var tStrikeDist float64
	for _, snake := range t.Snakes {
		if len(snake.Body) < len(ts.Body) {
			tStrikeDist += math.Abs(float64(snake.Body[0].X - ts.Body[0].X))
			tStrikeDist += math.Abs(float64(snake.Body[0].Y - ts.Body[0].Y))
		} else {
			tStrikeDist -= math.Abs(float64(snake.Body[0].X - ts.Body[0].X))
			tStrikeDist -= math.Abs(float64(snake.Body[0].Y - ts.Body[0].Y))
		}
	}

	pTailDist := math.Abs(float64(ps.Body[0].X-ps.Body[len(ps.Body)-1].X)) + math.Abs(float64(ps.Body[0].Y-ps.Body[len(ps.Body)-1].Y))
	tTailDist := math.Abs(float64(ts.Body[0].X-ts.Body[len(ts.Body)-1].X)) + math.Abs(float64(ts.Body[0].Y-ts.Body[len(ts.Body)-1].Y))
	if tTailDist <= pTailDist {
		res += chasingTail
	}

	if tStrikeDist <= pStrikeDist {
		res += betterStrikeDist
	} else {
		res += worseStrikeDist
	}

	if len(ps.Body) < len(ts.Body) {
		if len(ts.Body) < longest || ps.Health <= 15 {
			res += eatWhenHungry
		} else {
			res += eatWhenHealthy
		}
	}

	return res, nil
}

const (
	eatWhenHealthy   = -1
	eatWhenHungry    = 100
	lostGame         = -1000
	worseStrikeDist  = -3
	betterStrikeDist = 3
	chasingTail      = 50
)

func allMoves(snakes []rules.Snake, moves [][]rules.SnakeMove) [][]rules.SnakeMove {
	if len(snakes) == 0 {
		return moves
	}
	var res [][]rules.SnakeMove
	for _, dir := range []string{"up", "down", "left", "right"} {
		if len(moves) < 4 {
			res = append(res, []rules.SnakeMove{{ID: snakes[0].ID, Move: dir}})
			continue
		}
		for i := range moves {
			turn := []rules.SnakeMove{{ID: snakes[0].ID, Move: dir}}
			for _, move := range moves[i] {
				turn = append(turn, move)
			}
			res = append(res, turn)
		}
	}
	return allMoves(snakes[1:], res)
}

func pretty(req GameRequest) string {
	grid := makeGrid(req)

	var res string
	for y := req.Board.Height - 1; y >= 0; y-- {
		res += "|"
		for x := int32(0); x < req.Board.Width; x++ {
			res += string(grid.cells[rules.Point{x, y}].content)
		}
		res += "|\n"
	}
	return res
}

func prettyTTL(req GameRequest) string {
	grid := makeGrid(req)

	var res string
	for y := req.Board.Height - 1; y >= 0; y-- {
		res += "|"
		for x := int32(0); x < req.Board.Width; x++ {
			if grid.cells[rules.Point{x, y}].ttl == 0 {
				res += "   |"
				continue
			}
			res += fmt.Sprintf("%03d", grid.cells[rules.Point{x, y}].ttl) + "|"
		}
		res += "\n"
	}
	return res
}

func nextPos(p rules.Point, dir string) rules.Point {
	switch dir {
	case "up":
		return rules.Point{p.X, p.Y + 1}
	case "down":
		return rules.Point{p.X, p.Y - 1}
	case "left":
		return rules.Point{p.X - 1, p.Y}
	case "right":
		return rules.Point{p.X + 1, p.Y}
	}
	return rules.Point{}
}

func makeGrid(req GameRequest) grid {
	grid := makeAnonymousGridFromResponse(req.Board)

	for i, coord := range req.You.Body {
		if i == 0 {
			grid.cells[coord] = cell{content: 'Y', ttl: len(req.You.Body) - i}
			continue
		}
		grid.cells[coord] = cell{content: 'y', ttl: len(req.You.Body) - i}
	}

	return grid
}

func makeAnonymousGrid(state rules.BoardState) grid {
	cells := make(map[rules.Point]cell)
	for x := int32(0); x < state.Width; x++ {
		for y := int32(0); y < state.Height; y++ {
			cells[rules.Point{x, y}] = cell{content: ' '}
		}
	}

	for _, coord := range state.Food {
		cells[coord] = cell{content: 'f'}
	}

	for _, snake := range state.Snakes {
		for i, coord := range snake.Body {
			if i == 0 {
				cells[coord] = cell{content: 'S', ttl: len(snake.Body) - i}
				continue
			}
			cells[coord] = cell{content: 's', ttl: len(snake.Body) - i}
		}
	}

	return grid{
		width:  int(state.Width),
		height: int(state.Height),
		cells:  cells,
	}
}

func makeAnonymousGridFromResponse(state commands.BoardResponse) grid {
	cells := make(map[rules.Point]cell)
	for x := int32(0); x < state.Width; x++ {
		for y := int32(0); y < state.Height; y++ {
			cells[rules.Point{x, y}] = cell{content: ' '}
		}
	}

	for _, coord := range state.Food {
		cells[pointFromCoord(coord)] = cell{content: 'f'}
	}

	for _, snake := range state.Snakes {
		for i, coord := range snake.Body {
			if i == 0 {
				cells[pointFromCoord(coord)] = cell{content: 'S', ttl: len(snake.Body) - i}
				continue
			}
			cells[pointFromCoord(coord)] = cell{content: 's', ttl: len(snake.Body) - i}
		}
	}

	return grid{
		width:  int(state.Width),
		height: int(state.Height),
		cells:  cells,
	}
}

// HandleEnd is called when a game your Battlesnake was playing has ended.
// It's purely for informational purposes, no response required.
func HandleEnd(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Nothing to respond with here
	fmt.Print("END\n")
}

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/start", HandleStart)
	http.HandleFunc("/move", HandleMove)
	http.HandleFunc("/end", HandleEnd)

	http.HandleFunc("/striker/", striker.HandleIndex)
	http.HandleFunc("/striker/start", striker.HandleStart)
	http.HandleFunc("/striker/move", striker.HandleMove)
	http.HandleFunc("/striker/end", striker.HandleEnd)

	fmt.Printf("Starting Battlesnake Server at http://0.0.0.0:%s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	log.Fatal().Err(err).Msg("crashed")
}

func isValid(g grid, p rules.Point) bool {
	v, ok := g.cells[p]
	if !ok {
		return false
	}

	if v.content != ' ' && v.content != 'f' {
		return false
	}

	return true
}

func floodFill(g grid, p rules.Point, visited map[rules.Point]struct{}, limit, found int) int {
	if visited == nil {
		visited = make(map[rules.Point]struct{})
	}

	if _, ok := visited[p]; ok {
		return 0
	}
	visited[p] = struct{}{}

	if limit-found <= 0 {
		return 0
	}

	if !isValid(g, p) {
		if g.cells[p].ttl < limit {
			// We can follow tails indefinitely
			return limit - found
		}
		return 0
	}

	sum := found + 1
	sum = sum + floodFill(g, rules.Point{p.X, p.Y - 1}, visited, limit, sum)
	sum = sum + floodFill(g, rules.Point{p.X, p.Y + 1}, visited, limit, sum)
	sum = sum + floodFill(g, rules.Point{p.X - 1, p.Y}, visited, limit, sum)
	sum = sum + floodFill(g, rules.Point{p.X + 1, p.Y}, visited, limit, sum)

	return sum
}
