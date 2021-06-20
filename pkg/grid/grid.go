package grid

import (
	"battlesnake/pkg/types"
	"sort"
)

type Grid struct {
	Height, Width int
	Cells         map[types.Point]Cell
}

type Snake struct {
	Length int
	Name   string
	Head   types.Point
}

type CellContent int

const (
	ContentTypeUnknown CellContent = 0
	ContentTypeEmpty   CellContent = 1
	ContentTypeSnake   CellContent = 2
	ContentTypeFood    CellContent = 3
	ContentTypeHazard  CellContent = 4
)

type Cell struct {
	Content CellContent
	SnakeID string
	TTL     int
}

func Snakes(board types.BoardState) []Snake {
	snakeLengths := make(map[string]int)
	var snakes = make([]Snake, len(board.Snakes))
	for i, snake := range board.Snakes {
		if snake.EliminatedCause != "" {
			continue
		}

		snakeLengths[snake.Name] = len(snake.Body)
		snakes[i].Name = snake.Name
		snakes[i].Length = len(snake.Body)
		snakes[i].Head = snake.Body[0]
	}
	sort.Slice(snakes, func(i, j int) bool {
		return snakeLengths[snakes[i].Name] > snakeLengths[snakes[j].Name]
	})
	return snakes
}

func Make(board types.BoardState) Grid {
	cells := make(map[types.Point]Cell)
	for x := 0; x < board.Width; x++ {
		for y := 0; y < board.Height; y++ {
			cells[types.Point{X: x, Y: y}] = Cell{Content: ContentTypeEmpty}
		}
	}

	for _, p := range board.Food {
		cells[p] = Cell{Content: ContentTypeFood}
	}

	for _, snake := range board.Snakes {
		if snake.EliminatedCause != "" {
			continue
		}
		for i, coord := range snake.Body {
			cells[coord] = Cell{Content: ContentTypeSnake, TTL: len(snake.Body) - i, SnakeID: snake.ID}
		}
	}

	return Grid{
		Width:  board.Width,
		Height: board.Height,
		Cells:  cells,
	}
}
