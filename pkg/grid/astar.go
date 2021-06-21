package grid

import (
	"battlesnake/pkg/types"
	"container/heap"
)

func FindPath(g *Grid, from, to types.Point) (nextDir types.MoveDir, distance int) {
	path := search(g, from, to)
	if len(path) < 2 {
		return types.MoveDirUnknown, 0
	}
	nextCell := path[1]

	if nextCell.Y > from.Y {
		nextDir = "up"
	} else if nextCell.Y < from.Y {
		nextDir = "down"
	} else if nextCell.X > from.X {
		nextDir = "right"
	} else {
		nextDir = "left"
	}

	return nextDir, len(path) - 1
}

func search(g *Grid, to, from types.Point) []types.Point {
	nm := make(map[types.Point]*aStarNode)
	nq := &priorityQueue{}
	heap.Init(nq)
	fromNode := getNode(nm, from)
	fromNode.open = true
	heap.Push(nq, fromNode)
	for {
		if nq.Len() == 0 {
			// There's no path, return found false.
			return nil
		}
		current := heap.Pop(nq).(*aStarNode)
		current.open = false
		current.closed = true

		if current.p == getNode(nm, to).p {
			// Found a path.
			n := current
			var res []types.Point
			for n != nil {
				res = append(res, n.p)
				n = n.parent
			}
			return res
		}

		for _, nghbr := range []types.Point{
			{Y: current.p.Y, X: current.p.X - 1},
			{Y: current.p.Y, X: current.p.X + 1},
			{Y: current.p.Y - 1, X: current.p.X},
			{Y: current.p.Y + 1, X: current.p.X},
		} {
			v, ok := g.Cells[nghbr]
			if !ok {
				continue
			}

			cost := current.cost
			switch v.Content {
			case ContentTypeEmpty:
				cost += 1
				break
			case ContentTypeFood:
				cost += 2
				break
			case ContentTypeAvoid:
				cost += float64(g.Height * g.Width)
				break
			default:
				if nghbr == to {
					break
				}
				continue
			}

			neighborNode := getNode(nm, nghbr)
			if cost < neighborNode.cost {
				if neighborNode.open {
					heap.Remove(nq, neighborNode.index)
				}
				neighborNode.open = false
				neighborNode.closed = false
			}
			if !neighborNode.open && !neighborNode.closed {
				neighborNode.cost = cost
				neighborNode.open = true
				neighborNode.priority = cost + types.Distance(nghbr, to)
				neighborNode.parent = current
				heap.Push(nq, neighborNode)
			}
		}
	}
}

type aStarNode struct {
	p        types.Point
	cost     float64
	priority float64
	parent   *aStarNode
	open     bool
	closed   bool
	index    int
}

func getNode(nm map[types.Point]*aStarNode, p types.Point) *aStarNode {
	n, ok := nm[p]
	if !ok {
		n = &aStarNode{p: p}
		nm[p] = n
	}
	return n
}

// priorityQueue implements heap.Interface and holds Nodes.
type priorityQueue []*aStarNode

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*aStarNode)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}
