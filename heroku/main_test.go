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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/alcortesm/sample"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPretty(t *testing.T) {
	g := goldie.New(t)
	err := filepath.Walk("testdata", func(p string, info fs.FileInfo, err error) error {
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
	}
	for _, tc := range tests {
		t.Run(tc.requestFile, func(t *testing.T) {
			file, err := os.Open(path.Join("testdata", tc.requestFile))
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
			actual := allMoves(tc.snakes, nil)
			require.Len(t, actual, tc.len)
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
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			request := parseGameRequest(t, path.Join("testdata", tc.requestFile))

			grid := makeGrid(request)

			actual := floodFill(grid, tc.start, nil, tc.limit)
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

func TestBruteForce(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	ruleset := &rules.SoloRuleset{
		StandardRuleset: rules.StandardRuleset{
			FoodSpawnChance: 15,
			MinimumFood:     1,
		},
	}

	results := make([]float64, 0, 1000)
	var worstScore int
	var worstState rules.BoardState
	for i := 0; i < 100; i++ {
		state, err := ruleset.CreateInitialBoardState(11, 11, []string{"Charmer"})
		require.NoError(t, err)

		var turn int
		for v := false; !v; v, _ = ruleset.IsGameOver(state) {
			turn++

			mr, err := doMove(makeGameRequest(turn, *state))
			require.NoError(t, err)

			state, err = ruleset.CreateNextBoardState(state, []rules.SnakeMove{{state.Snakes[0].ID, mr.best()}})
			require.NoError(t, err)
		}
		results = append(results, float64(turn))

		if worstScore == 0 || turn < worstScore {
			worstState = *state
		}
	}

	assert.Len(t, results, 100)

	mean, err := sample.Mean(results)
	require.NoError(t, err)
	assert.Greater(t, mean, 720.38)

	sd, err := sample.StandardDeviation(results)
	require.NoError(t, err)
	assert.Less(t, sd, mean/2.0)

	fmt.Println(mean, sd)

	goldie.New(t).Update(t, t.Name()+time.Now().Format("/2006-01-02-15-04-05"), []byte(pretty(makeGameRequest(worstScore, worstState))))
}

func makeGameRequest(turn int, state rules.BoardState) GameRequest {
	return GameRequest{
		Turn:  turn,
		Board: state,
		You: Battlesnake{
			ID:     state.Snakes[0].ID,
			Health: state.Snakes[0].Health,
			Body:   state.Snakes[0].Body,
			Head:   state.Snakes[0].Body[0],
			Length: int32(len(state.Snakes[0].Body)),
		},
	}
}
