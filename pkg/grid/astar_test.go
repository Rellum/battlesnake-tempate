package grid

import (
	"battlesnake/pkg/types"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	g := goldie.New(t)

	tl := []struct {
		name        string
		requestFile string
		from, to    types.Point
	}{
		{
			name:        "circle_back",
			requestFile: "case_13.json",
			from:        types.Point{X: 3, Y: 1},
			to:          types.Point{X: 1, Y: 1},
		},
		{
			name:        "direct",
			requestFile: "case_8.json",
			from:        types.Point{X: 4, Y: 10},
			to:          types.Point{X: 6, Y: 10},
		},
		{
			name:        "food",
			requestFile: "case_5.json",
			from:        types.Point{X: 9, Y: 10},
			to:          types.Point{X: 10, Y: 4},
		},
	}
	for _, tc := range tl {
		t.Run(tc.name, func(t *testing.T) {
			request := parseGameRequest(t, filepath.Join("../../testdata", tc.requestFile))
			grd := Make(request.Board)

			actual := search(&grd, tc.from, tc.to)

			g.Update(t, tc.name, printPath(grd.Height, grd.Width, actual))
		})
	}
}

func TestFindPath(t *testing.T) {
	tl := []struct {
		name        string
		requestFile string
		from, to    types.Point
		dir         types.MoveDir
		dist        int
	}{
		{
			name:        "circle_back",
			requestFile: "case_13.json",
			from:        types.Point{X: 3, Y: 1},
			to:          types.Point{X: 1, Y: 1},
			dir:         types.MoveDirUp,
			dist:        4,
		},
		{
			name:        "direct",
			requestFile: "case_8.json",
			from:        types.Point{X: 4, Y: 10},
			to:          types.Point{X: 6, Y: 10},
			dir:         types.MoveDirRight,
			dist:        2,
		},
		{
			name:        "food",
			requestFile: "case_5.json",
			from:        types.Point{X: 9, Y: 10},
			to:          types.Point{X: 10, Y: 4},
			dir:         types.MoveDirRight,
			dist:        7,
		},
	}
	for _, tc := range tl {
		t.Run(tc.name, func(t *testing.T) {
			request := parseGameRequest(t, filepath.Join("../../testdata", tc.requestFile))
			grd := Make(request.Board)

			dir, dist := FindPath(&grd, tc.from, tc.to)
			require.Equal(t, tc.dir, dir)
			require.Equal(t, tc.dist, dist)
		})
	}
}

func printPath(height, width int, pl []types.Point) []byte {
	pm := make(map[types.Point]int)
	for i, p := range pl {
		pm[p] = i
	}

	var res bytes.Buffer
	for y := height - 1; y >= 0; y-- {
		fmt.Fprint(&res, "|")
		for x := 0; x < width; x++ {
			if i, ok := pm[types.Point{x, y}]; ok {
				fmt.Fprintf(&res, "%d", i)
			} else {
				fmt.Fprint(&res, " ")
			}
		}
		fmt.Fprintln(&res, "|")
	}
	return res.Bytes()
}

func parseGameRequest(t *testing.T, filename string) types.GameRequest {
	b, err := os.ReadFile(filename)
	require.NoError(t, err)

	var request types.GameRequest
	err = json.Unmarshal(b, &request)
	require.NoError(t, err)

	return request
}
