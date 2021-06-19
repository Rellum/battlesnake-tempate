package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeGraph(t *testing.T) {
	g := makeGrid(parseGameRequest(t, path.Join("testdata", "case_1.json")))
	assert.Equal(t, g.width*g.height, len(g.cells))
	actual := makeGraph(g)
	assert.Equal(t, g.width*(g.height-1)*2+(g.width-1)*g.height*2-2-3-3-4-4-4-4, len(actual.edges))
}
