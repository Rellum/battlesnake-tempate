package main

import (
	"github.com/BattlesnakeOfficial/rules"
)

func makeGraph(g grid) graph {
	res := graph{edges: make(map[edge]struct{})}
	for p, c := range g.cells {
		if c.content != ' ' && c.content != 'f' {
			continue
		}

		if p.X > 0 {
			res.edges[edge{to: p, from: rules.Point{Y: p.Y, X: p.X - 1}}] = struct{}{}
		}

		if p.Y > 0 {
			res.edges[edge{to: p, from: rules.Point{Y: p.Y - 1, X: p.X}}] = struct{}{}
		}

		if p.X < int32(g.width)-1 {
			res.edges[edge{to: p, from: rules.Point{Y: p.Y, X: p.X + 1}}] = struct{}{}
		}

		if p.Y < int32(g.height)-1 {
			res.edges[edge{to: p, from: rules.Point{Y: p.Y + 1, X: p.X}}] = struct{}{}
		}
	}
	return res
}
