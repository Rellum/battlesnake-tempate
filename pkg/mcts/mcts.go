package mcts

import (
	"battlesnake/pkg/grid"
	"battlesnake/pkg/types"
	"context"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/0xhexnumbers/gmcts/v2"
	"github.com/BattlesnakeOfficial/rules"
)

var allDirections = []string{"up", "down", "left", "right"}

func Search(ctx context.Context, s rules.BoardState, hazards []rules.Point, me string, turn int32) (string, error) {
	sort.Slice(s.Snakes, func(i, j int) bool {
		return s.Snakes[i].ID == me
	})

	grid := grid.MakeFromRulesState(s, hazards)
	gameState := &game{
		rs:             inferRuleset(&s, hazards, turn),
		state:          s,
		me:             me,
		players:        make(map[string]gmcts.Player),
		remainingTurns: 60,
		grid:           &grid,
		mu:             new(sync.Mutex),
	}

	for i, snake := range s.Snakes {
		gameState.players[snake.ID] = gmcts.Player(i)
		if snake.EliminatedCause != "" {
			continue
		}
		gameState.moves = append(gameState.moves, rules.SnakeMove{ID: snake.ID})
	}

	mcts := gmcts.NewMCTS(gameState)

	var wait sync.WaitGroup
	concurrentTrees := runtime.NumCPU()
	concurrentTrees = 1
	wait.Add(concurrentTrees)
	for i := 0; i < concurrentTrees; i++ {
		go func() {
			tree := mcts.SpawnTree()
			tree.SearchContext(ctx)
			mcts.AddTree(tree)
			wait.Done()
		}()
	}
	wait.Wait()

	bestAction, err := mcts.BestAction()
	if err != nil {
		return "", err
	}

	return string(gameState.availableActions[bestAction]), nil
}

func inferRuleset(s *rules.BoardState, hazards []rules.Point, turn int32) rules.Ruleset {
	stdRules := rules.StandardRuleset{
		FoodSpawnChance: 15,
		MinimumFood:     1,
	}

	if len(s.Snakes) == 1 {
		return &rules.SoloRuleset{stdRules}
	}

	if len(hazards) > 0 {
		return &rules.RoyaleRuleset{
			StandardRuleset:   stdRules,
			Seed:              time.Now().UTC().UnixNano(),
			Turn:              turn,
			ShrinkEveryNTurns: 10,
			DamagePerTurn:     1,
			OutOfBounds:       append([]rules.Point{}, hazards...),
		}
	}

	return &stdRules
}

type game struct {
	rs               rules.Ruleset
	state            rules.BoardState
	grid             *grid.Grid
	me               string
	moves            []rules.SnakeMove
	players          map[string]gmcts.Player
	remainingTurns   int32
	availableActions []types.MoveDir
	mu               *sync.Mutex
}

func (g *game) Len() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.len()
}

func (g *game) len() int {
	g.ensureActions()
	return len(g.availableActions)
}

func (g *game) ensureActions() {
	if g.availableActions != nil {
		return
	}

	snakeID := nextPlayer(g.moves)
	if snakeID == "" {
		return
	}

	var h types.Point
	for _, snake := range g.state.Snakes {
		if snake.ID != snakeID {
			continue
		}
		h = types.Point{X: int(snake.Body[0].X), Y: int(snake.Body[0].Y)}
		break
	}

	for _, nghbr := range []struct {
		p   types.Point
		dir types.MoveDir
	}{
		{p: types.Point{Y: h.Y, X: h.X - 1}, dir: types.MoveDirLeft},
		{p: types.Point{Y: h.Y, X: h.X + 1}, dir: types.MoveDirRight},
		{p: types.Point{Y: h.Y - 1, X: h.X}, dir: types.MoveDirDown},
		{p: types.Point{Y: h.Y + 1, X: h.X}, dir: types.MoveDirUp},
	} {
		if grid.IsValid(g.grid, nghbr.p) {
			g.availableActions = append(g.availableActions, nghbr.dir)
		}
	}
}

func (g *game) ApplyAction(actionID int) (gmcts.Game, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ensureActions()
	moves := append([]rules.SnakeMove{}, g.moves...)
	for i := range moves {
		if moves[i].Move != "" {
			continue
		}
		moves[i].Move = string(g.availableActions[actionID])
		break
	}

	res := game{
		rs:      g.rs,
		state:   g.state,
		me:      g.me,
		players: g.players,
		mu:      new(sync.Mutex),
	}

	if snakeID := nextPlayer(moves); snakeID != "" {
		res.moves = moves
		res.remainingTurns = g.remainingTurns
		res.grid = g.grid
		return &res, nil
	}

	state, err := g.rs.CreateNextBoardState(&g.state, moves)
	if err != nil {
		return nil, err
	}
	res.state = *state
	res.remainingTurns = g.remainingTurns - 1

	var hazards []rules.Point
	if rs, ok := g.rs.(*rules.RoyaleRuleset); ok {
		hazards = append([]rules.Point{}, rs.OutOfBounds...)
	}

	grid := grid.MakeFromRulesState(*state, hazards)
	res.grid = &grid
	for _, snake := range state.Snakes {
		if snake.EliminatedCause != "" {
			continue
		}
		res.moves = append(res.moves, rules.SnakeMove{ID: snake.ID})
	}

	return &res, nil
}

func nextPlayer(moves []rules.SnakeMove) string {
	for _, move := range moves {
		if move.Move == "" {
			return move.ID
		}
	}
	return ""
}

func (g *game) Hash() interface{} {
	return rand.Int63()
}

func (g *game) Player() gmcts.Player {
	g.mu.Lock()
	defer g.mu.Unlock()
	return getPlayer(g.players, nextPlayer(g.moves))
}

func (g *game) IsTerminal() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.remainingTurns <= 0 {
		return true
	}

	if g.len() == 0 {
		return true
	}

	over, err := g.rs.IsGameOver(&g.state)
	if err != nil {
		panic(err)
	}

	if over {
		return true
	}

	for _, snake := range g.state.Snakes {
		if snake.ID != g.me {
			continue
		}
		return snake.EliminatedCause != ""
	}

	panic("BUG: isTerminal did not find my state")
}

func (g *game) Winners() []gmcts.Player {
	g.mu.Lock()
	defer g.mu.Unlock()
	res := []gmcts.Player{gmcts.Player(len(g.state.Snakes))}
	for _, snake := range g.state.Snakes {
		if snake.EliminatedCause != "" {
			continue
		}

		if snake.ID == nextPlayer(g.moves) && g.len() == 0 {
			continue
		}

		res = append(res, getPlayer(g.players, snake.ID))
	}

	return res
}

func getPlayer(players map[string]gmcts.Player, snakeID string) gmcts.Player {
	res, ok := players[snakeID]
	if !ok {
		panic("BUG: player not found")
	}
	return res
}
