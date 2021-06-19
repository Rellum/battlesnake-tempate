package grid

import (
	"battlesnake/pkg/types"
)

type Grid struct {
	Height, Width int
	Cells         map[types.Point]Cell
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
