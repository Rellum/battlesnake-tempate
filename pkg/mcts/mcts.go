package mcts

import (
	"context"
	"sort"
	"time"

	"github.com/0xhexnumbers/gmcts/v2"
	"github.com/BattlesnakeOfficial/rules"
)

var allDirections = []string{"up", "down", "left", "right"}

func Search(ctx context.Context, s rules.BoardState, hazards []rules.Point, me string, turn int32) (string, error) {
	sort.Slice(s.Snakes, func(i, j int) bool {
		return s.Snakes[i].ID == me
	})

	gameState := &game{
		rs:      inferRuleset(&s, hazards, turn),
		state:   s,
		me:      me,
		players: make(map[string]gmcts.Player),
	}

	for i, snake := range s.Snakes {
		gameState.players[snake.ID] = gmcts.Player(i)
		if snake.EliminatedCause != "" {
			continue
		}
		gameState.moves = append(gameState.moves, rules.SnakeMove{ID: snake.ID})
	}

	mcts := gmcts.NewMCTS(gameState)

	tree := mcts.SpawnTree()
	tree.SearchContext(ctx)

	mcts.AddTree(tree)

	bestAction, err := mcts.BestAction()
	if err != nil {
		return "", err
	}

	return allDirections[bestAction], nil
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
		}
	}

	return &stdRules
}

type game struct {
	rs      rules.Ruleset
	state   rules.BoardState
	me      string
	moves   []rules.SnakeMove
	players map[string]gmcts.Player
}

func (g *game) Len() int {
	return 4
}

func (g *game) ApplyAction(actionID int) (gmcts.Game, error) {
	moves := append([]rules.SnakeMove{}, g.moves...)
	for i := range moves {
		if moves[i].Move != "" {
			continue
		}
		moves[i].Move = allDirections[actionID]
		break
	}

	res := game{
		rs:      g.rs,
		state:   g.state,
		me:      g.me,
		players: g.players,
	}

	if snakeID := nextPlayer(moves); snakeID != "" {
		res.moves = moves
		return &res, nil
	}

	state, err := g.rs.CreateNextBoardState(&g.state, moves)
	if err != nil {
		return nil, err
	}
	res.state = *state
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
	return g
}

func (g *game) Player() gmcts.Player {
	return getPlayer(g.players, nextPlayer(g.moves))
}

func (g *game) IsTerminal() bool {
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
	var res []gmcts.Player
	for _, snake := range g.state.Snakes {
		if snake.EliminatedCause != "" {
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
