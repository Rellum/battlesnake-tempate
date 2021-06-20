package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/cli/commands"
	"github.com/alcortesm/sample"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPretty(t *testing.T) {
	g := goldie.New(t)
	err := filepath.Walk("../testdata", func(p string, info fs.FileInfo, err error) error {
		if filepath.Ext(info.Name()) != ".json" {
			return nil
		}
		t.Run(info.Name(), func(t *testing.T) {
			request := parseGameRequest(t, p)

			g.Update(t, strings.TrimSuffix(info.Name(), ".json"), []byte(pretty(request)))
		})
		return nil
	})
	require.NoError(t, err)
}

func TestTTL(t *testing.T) {
	g := goldie.New(t)
	err := filepath.Walk("../testdata", func(p string, info fs.FileInfo, err error) error {
		if filepath.Ext(info.Name()) != ".json" {
			return nil
		}
		t.Run(info.Name(), func(t *testing.T) {
			request := parseGameRequest(t, p)

			g.Update(t, strings.TrimSuffix(info.Name(), ".json")+"_ttl", []byte(prettyTTL(request)))
		})
		return nil
	})
	require.NoError(t, err)
}

func TestMove(t *testing.T) {
	tests := []struct {
		requestFile string
		possible    []string
	}{
		{
			requestFile: "case_1.json",
			possible:    []string{"up", "down", "left", "right"},
		},
		{
			requestFile: "case_2.json",
			possible:    []string{"left", "right"},
		},
		{
			requestFile: "case_3.json",
			possible:    []string{"up", "right"},
		},
		{
			requestFile: "case_4.json",
			possible:    []string{"down"},
		},
		{
			requestFile: "case_5.json",
			possible:    []string{"left", "right"},
		},
		{
			requestFile: "case_6.json",
			possible:    []string{"left", "right"},
		},
		{
			requestFile: "case_7.json",
			possible:    []string{"left", "right"},
		},
		{
			requestFile: "case_8.json",
			possible:    []string{"right"},
		},
		{
			requestFile: "case_9.json",
			possible:    []string{"up"},
		},
		{
			requestFile: "case_10.json",
			possible:    []string{"left", "right"},
		},
		{
			requestFile: "case_11.json",
			possible:    []string{"down"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.requestFile, func(t *testing.T) {
			file, err := os.Open(path.Join("../testdata", tc.requestFile))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			HandleMove(recorder, httptest.NewRequest(http.MethodPost, "/move", file))

			var response MoveResponse
			err = json.NewDecoder(recorder.Body).Decode(&response)
			require.NoError(t, err)

			require.Contains(t, tc.possible, response.Move)
		})
	}
}

func TestDoMove(t *testing.T) {
	assertPos := func(i int, dir string) func(*testing.T, []scoredMove) {
		return func(t *testing.T, r []scoredMove) {
			require.Equal(t, dir, r[i-1].dir)
		}
	}

	tests := []struct {
		requestFile string
		assertions  []func(*testing.T, []scoredMove)
	}{
		{
			requestFile: "case_1.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(1, "up"),
			},
		},
		{
			requestFile: "case_2.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(3, "up"),
				assertPos(4, "down"),
			},
		},
		{
			requestFile: "case_3.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(3, "left"),
				assertPos(4, "down"),
			},
		},
		{
			requestFile: "case_4.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(1, "down"),
				assertPos(4, "right"),
			},
		},
		{
			requestFile: "case_5.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(3, "up"),
				assertPos(4, "down"),
			},
		},
		{
			requestFile: "case_6.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(3, "up"),
				assertPos(4, "down"),
			},
		},
		{
			requestFile: "case_7.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(3, "up"),
				assertPos(4, "down"),
			},
		},
		{
			requestFile: "case_8.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(1, "right"),
				assertPos(2, "left"),
				assertPos(3, "down"),
				assertPos(4, "up"),
			},
		},
		{
			requestFile: "case_9.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(1, "up"),
				assertPos(2, "down"),
				assertPos(3, "left"),
				assertPos(4, "right"),
			},
		},
		{
			requestFile: "case_10.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(3, "up"),
				assertPos(4, "down"),
			},
		},
		{
			requestFile: "case_11.json",
			assertions: []func(*testing.T, []scoredMove){
				assertPos(1, "down"),
				assertPos(2, "left"),
				assertPos(3, "up"),
				assertPos(4, "right"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.requestFile, func(t *testing.T) {
			b, err := os.ReadFile(path.Join("../testdata", tc.requestFile))
			require.NoError(t, err)

			request := GameRequest{}
			err = json.Unmarshal(b, &request)
			require.NoError(t, err)

			actual, err := doMove(request)
			require.NoError(t, err)

			result := []scoredMove{
				{dir: "up", score: actual.u},
				{dir: "down", score: actual.d},
				{dir: "left", score: actual.l},
				{dir: "right", score: actual.r},
			}
			sort.Slice(result, func(i, j int) bool {
				return result[i].score > result[j].score
			})

			for _, assertion := range tc.assertions {
				assertion(t, result)
			}
		})
	}
}

type scoredMove struct {
	dir   string
	score int
}

func TestAllMoves(t *testing.T) {
	tl := []struct {
		name   string
		snakes []rules.Snake
		len    int
	}{
		{
			name:   "single",
			snakes: []rules.Snake{{ID: "1"}},
			len:    4,
		},
		{
			name:   "multiple",
			snakes: []rules.Snake{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			len:    64,
		},
	}
	for _, tc := range tl {
		t.Run(tc.name, func(t *testing.T) {
			count := make(map[rules.SnakeMove]struct{})
			actual := allMoves(tc.snakes, nil)
			require.Len(t, actual, tc.len)
			for i := range actual {
				require.Len(t, actual[i], len(tc.snakes))
				for j := range actual[i] {
					count[actual[i][j]] = struct{}{}
				}
			}
			require.Len(t, count, len(tc.snakes)*4)
		})
	}
}

func TestFloodFill(t *testing.T) {
	tests := []struct {
		requestFile string
		start       rules.Point
		limit       int
		expected    int
	}{
		{
			requestFile: "case_7.json",
			start:       rules.Point{X: 2, Y: 10},
			limit:       1000,
			expected:    5,
		},
		{
			requestFile: "case_7.json",
			start:       rules.Point{X: 4, Y: 10},
			limit:       1000,
			expected:    109,
		},
		{
			requestFile: "case_7.json",
			start:       rules.Point{X: 3, Y: 10},
			limit:       1000,
		},
		{
			requestFile: "case_7.json",
			start:       rules.Point{X: 2, Y: 10},
			limit:       4,
			expected:    4,
		},
		{
			requestFile: "case_7.json",
			start:       rules.Point{X: 4, Y: 10},
			limit:       4,
			expected:    4,
		},
		{
			requestFile: "case_7.json",
			start:       rules.Point{X: 3, Y: 10},
			limit:       4,
		},
		{
			requestFile: "case_9.json",
			start:       rules.Point{X: 0, Y: 4},
			limit:       12,
			expected:    10,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s_x%d_y%d_l%d", tc.requestFile, tc.start.X, tc.start.Y, tc.limit), func(t *testing.T) {
			request := parseGameRequest(t, path.Join("../testdata", tc.requestFile))

			grid := makeGrid(request)

			visited := make(map[rules.Point]struct{})
			actual := floodFill(grid, tc.start, visited, tc.limit, 0)

			var visual string
			for y := request.Board.Height - 1; y >= 0; y-- {
				visual += "|"
				for x := int32(0); x < request.Board.Width; x++ {
					if _, ok := visited[rules.Point{x, y}]; ok {
						visual += "v"
					} else {
						visual += " "
					}
				}
				visual += "|\n"
			}
			t.Log(visual)

			require.Equal(t, tc.expected, actual)
		})
	}
}

func parseGameRequest(t *testing.T, filename string) GameRequest {
	b, err := os.ReadFile(filename)
	assert.NoError(t, err)

	var request GameRequest
	err = json.Unmarshal(b, &request)
	assert.NoError(t, err)

	return request
}

func BenchmarkBruteForce(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	ruleset := &rules.SoloRuleset{
		StandardRuleset: rules.StandardRuleset{
			FoodSpawnChance: 15,
			MinimumFood:     1,
		},
	}

	results := make([]float64, 0, 1000)
	for i := 0; i < b.N; i++ {
		state, err := ruleset.CreateInitialBoardState(11, 11, []string{"Charmer"})
		require.NoError(b, err)

		var turn int32
		for v := false; !v; v, _ = ruleset.IsGameOver(state) {
			turn++

			mr, err := doMove(makeGameRequest(turn, *state))
			require.NoError(b, err)

			state, err = ruleset.CreateNextBoardState(state, []rules.SnakeMove{{state.Snakes[0].ID, mr.best()}})
			require.NoError(b, err)
		}
		results = append(results, float64(turn))
	}

	assert.Len(b, results, b.N)

	if b.N == 1 {
		return
	}

	mean, err := sample.Mean(results)
	require.NoError(b, err)
	//assert.Greater(b, mean, 1080.33)

	sd, err := sample.StandardDeviation(results)
	require.NoError(b, err)
	assert.Less(b, sd, mean/3.0)

	b.Log(b.N, mean, sd)
}

func makeGameRequest(turn int32, state rules.BoardState) GameRequest {
	return GameRequest{
		Turn: turn,
		Board: commands.BoardResponse{
			Height: state.Height,
			Width:  state.Width,
			Food:   coordFromPointSlice(state.Food),
			Snakes: buildSnakesResponse(state.Snakes),
		},
		You: Battlesnake{
			ID:     state.Snakes[0].ID,
			Health: state.Snakes[0].Health,
			Body:   state.Snakes[0].Body,
			Head:   state.Snakes[0].Body[0],
			Length: int32(len(state.Snakes[0].Body)),
		},
	}
}
