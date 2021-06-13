package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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

//go:embed info.json

var snakeInfo []byte

// HandleIndex is called when your Battlesnake is created and refreshed
// by play.battlesnake.com. BattlesnakeInfoResponse contains information about
// your Battlesnake, including what it should look like on the game board.
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(snakeInfo)
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

	log.Debug().Msg(pretty(request))

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
	ruleset := rules.SoloRuleset{rules.StandardRuleset{
		FoodSpawnChance: 15,
		MinimumFood:     1,
	}}

	bestDir, _, err := simulate(&ruleset, &request.Board, request.You.ID, 4)
	if err != nil {
		return nil, err
	}

	if bestDir == "" {
		grid := makeAnonymousGrid(request.Board)
		for _, dir := range []string{"up", "down", "left", "right"} {
			if isValid(grid, nextPos(request.You.Head, dir)) {
				bestDir = dir
				break
			}
		}
	}

	var res moveResult
	switch bestDir {
	case "up":
		res.u = 1
	case "down":
		res.d = 1
	case "left":
		res.l = 1
	case "right":
		res.r = 1
	}
	return &res, nil
}

func simulate(rs rules.Ruleset, state *rules.BoardState, me string, depth int) (string, int, error) {
	if depth == 0 {
		return "", 0, nil
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
	var max int
	var best string
	var eg errgroup.Group
	for _, ml := range allMoves(state.Snakes, nil) {
		ml := ml
		eg.Go(func() error {
			for _, move := range ml {
				if floodFill(grid, nextPos(snakes[move.ID].head, move.Move), nil, snakes[move.ID].len) < snakes[move.ID].len {
					return nil
				}
			}

			newState, err := rs.CreateNextBoardState(state, ml)
			if err != nil {
				log.Fatal().Err(err).Stack().Msg("")
			}

			score, err := scoreMove(rs, state, newState, me)
			if err != nil {
				return err
			}

			_, i, err := simulate(rs, newState, me, depth-1)
			if err != nil {
				return err
			}

			i += score
			var dir string
			for _, move := range ml {
				if move.ID == me {
					dir = move.Move
					break
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if max < i {
				best = dir
				max = i
			}
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return "", 0, err
	}
	return best, max, nil
}

func scoreMove(rs rules.Ruleset, p, t *rules.BoardState, me string) (int, error) {
	over, err := rs.IsGameOver(t)
	if err != nil {
		return 0, err
	}
	if over {
		return 0, nil
	}

	var ps rules.Snake
	for _, snake := range p.Snakes {
		if snake.ID != me {
			continue
		}
		ps = snake
	}

	if ps.Health < 30 {
		return 1, nil
	}

	var ts rules.Snake
	for _, snake := range t.Snakes {
		if snake.ID != me {
			continue
		}
		ts = snake
	}

	if len(ps.Body) < len(ts.Body) {
		return 1, nil
	}

	return 2, nil
}

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
			res += string(grid[rules.Point{x, y}])
		}
		res += "|\n"
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
	grid := makeAnonymousGrid(req.Board)

	for i, coord := range req.You.Body {
		if i == 0 {
			grid[coord] = 'Y'
			continue
		}
		grid[coord] = 'y'
	}

	return grid
}

func makeAnonymousGrid(state rules.BoardState) grid {
	grid := make(map[rules.Point]rune)
	for x := int32(0); x < state.Width; x++ {
		for y := int32(0); y < state.Height; y++ {
			grid[rules.Point{x, y}] = ' '
		}
	}

	for _, coord := range state.Food {
		grid[coord] = 'f'
	}

	for _, snake := range state.Snakes {
		for i, coord := range snake.Body {
			if i == 0 {
				grid[coord] = 'S'
				continue
			}
			grid[coord] = 's'
		}
	}

	return grid
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

	fmt.Printf("Starting Battlesnake Server at http://0.0.0.0:%s...\n", port)
	log.Fatal().Err(http.ListenAndServe(":"+port, nil)).Msg("")
}

func isValid(g grid, p rules.Point) bool {
	if g[p] != ' ' && g[p] != 'f' {
		return false
	}

	return true
}

func floodFill(g grid, p rules.Point, visited map[rules.Point]struct{}, limit int) int {
	if limit == 0 {
		return 0
	}

	if visited == nil {
		visited = make(map[rules.Point]struct{})
	}

	if _, ok := visited[p]; ok {
		return 0
	}
	visited[p] = struct{}{}

	if !isValid(g, p) {
		return 0
	}

	sum := 1
	sum = sum + floodFill(g, rules.Point{p.X, p.Y - 1}, visited, limit-sum)
	sum = sum + floodFill(g, rules.Point{p.X, p.Y + 1}, visited, limit-sum)
	sum = sum + floodFill(g, rules.Point{p.X - 1, p.Y}, visited, limit-sum)
	sum = sum + floodFill(g, rules.Point{p.X + 1, p.Y}, visited, limit-sum)

	return sum
}
