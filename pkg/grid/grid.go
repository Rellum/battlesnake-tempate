package grid

import (
	"battlesnake/pkg/types"
	"bytes"
	"fmt"
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

func IsValid(g *Grid, p types.Point) bool {
	v, ok := g.Cells[p]
	if !ok {
		return false
	}

	if v.Content == ContentTypeEmpty || v.Content == ContentTypeFood {
		return true
	}

	return false
}

func FloodFill(g *Grid, p types.Point, limit int) int {
	return floodFill(g, p, make(map[types.Point]struct{}), limit, 0)
}

func floodFill(g *Grid, p types.Point, visited map[types.Point]struct{}, limit, found int) int {
	if _, ok := visited[p]; ok {
		return 0
	}
	visited[p] = struct{}{}

	if limit-found <= 0 {
		return 0
	}

	if !IsValid(g, p) {
		return 0
	}

	sum := found + 1
	sum = sum + floodFill(g, types.Point{X: p.X, Y: p.Y - 1}, visited, limit, sum)
	sum = sum + floodFill(g, types.Point{X: p.X, Y: p.Y + 1}, visited, limit, sum)
	sum = sum + floodFill(g, types.Point{X: p.X - 1, Y: p.Y}, visited, limit, sum)
	sum = sum + floodFill(g, types.Point{X: p.X + 1, Y: p.Y}, visited, limit, sum)

	return sum
}

func Print(g *Grid) []byte {
	var res bytes.Buffer
	for y := g.Height - 1; y >= 0; y-- {
		fmt.Fprint(&res, "|")
		for x := 0; x < g.Width; x++ {
			c := g.Cells[types.Point{x, y}]
			switch c.Content {
			case ContentTypeFood:
				fmt.Fprint(&res, "*")
			case ContentTypeEmpty:
				fmt.Fprint(&res, " ")
			case ContentTypeSnake:
				fmt.Fprint(&res, "s")
			}
		}
		fmt.Fprintln(&res, "|")
	}
	return res.Bytes()
}

func PrintTTL(g *Grid) []byte {
	var res bytes.Buffer
	for y := g.Height - 1; y >= 0; y-- {
		fmt.Fprint(&res, "|")
		for x := 0; x < g.Width; x++ {
			if g.Cells[types.Point{x, y}].TTL == 0 {
				fmt.Fprint(&res, "  |")
				continue
			}
			fmt.Fprintf(&res, "%02d|", g.Cells[types.Point{x, y}].TTL)
		}
		fmt.Fprintln(&res, "|")
	}
	return res.Bytes()
}
