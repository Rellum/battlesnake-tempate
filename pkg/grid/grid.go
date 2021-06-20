package grid

import (
	"battlesnake/pkg/types"
	"sort"

	"github.com/beefsack/go-astar"
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

func FindPath(g *Graph, from, to types.Point) (nextDir types.MoveDir, distance float64, found bool) {
	path, distance, found := astar.Path(makeNode(g, to), makeNode(g, from))
	if !found {
		return types.MoveDirUnknown, 0, false
	}
	nextCell := path[1].(*node).Point

	if nextCell.Y > from.Y {
		nextDir = "up"
	} else if nextCell.Y < from.Y {
		nextDir = "down"
	} else if nextCell.X > from.X {
		nextDir = "right"
	} else {
		nextDir = "left"
	}

	return nextDir, distance, found
}

func MakeGraph(g Grid) *Graph {
	return &Graph{
		previous: make(map[types.Point]astar.Pather),
		grid:     &g,
	}
}

func makeNode(g *Graph, p types.Point) *node {
	var res = node{
		Point: p,
		graph: g,
	}
	g.previous[p] = &res
	return &res
}

type Graph struct {
	previous map[types.Point]astar.Pather
	grid     *Grid
}

type node struct {
	types.Point
	graph *Graph
}

func (n *node) PathNeighbors() []astar.Pather {
	var res []astar.Pather
	for _, nghbr := range []types.Point{
		{Y: n.Y, X: n.X - 1},
		{Y: n.Y, X: n.X + 1},
		{Y: n.Y - 1, X: n.X},
		{Y: n.Y + 1, X: n.X},
	} {
		pthr, ok := n.graph.previous[nghbr]
		if ok {
			res = append(res, pthr)
			continue
		}

		if n.graph.grid.Cells[nghbr].Content != ContentTypeEmpty && n.graph.grid.Cells[nghbr].Content != ContentTypeFood {
			continue
		}

		pthr = &node{graph: n.graph, Point: nghbr}
		n.graph.previous[nghbr] = pthr
		res = append(res, pthr)
	}

	return res
}

func (n *node) PathNeighborCost(astar.Pather) float64 {
	return 1
}

func (n *node) PathEstimatedCost(toPather astar.Pather) float64 {
	return types.Distance(n.Point, toPather.(*node).Point)
}
