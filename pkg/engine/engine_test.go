package engine

import (
	"battlesnake/pkg/types"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	g := goldie.New(t)

	tl := []struct {
		name  string
		moves []types.SnakeMove
	}{
		{
			name: "no_moves",
		},
		{
			name: "some_valid_moves",
			moves: []types.SnakeMove{{
				ID:    "snake-b67f4906-94ae-11ea-bb37",
				Move:  types.MoveDirLeft,
				Shout: "Something!",
			}},
		},
		{
			name: "eating",
			moves: []types.SnakeMove{{
				ID:    "snake-b67f4906-94ae-11ea-bb37",
				Move:  types.MoveDirUp,
				Shout: "Something else!",
			}},
		},
	}
	for _, tc := range tl {
		t.Run(tc.name, func(t *testing.T) {

			request := parseGameRequest(t, filepath.Join("../../testdata", "case_1.json"))
			//g.Update(t, "case_1", types.Print(request.Board, nil))

			actual := MoveSnakes(request.Board, tc.moves)

			g.Update(t, tc.name, types.Print(actual, nil))
			g.Assert(t, "case_1", types.Print(request.Board, nil))
		})
	}
}

func parseGameRequest(t *testing.T, filename string) types.GameRequest {
	b, err := os.ReadFile(filename)
	require.NoError(t, err)

	var request types.GameRequest
	err = json.Unmarshal(b, &request)
	require.NoError(t, err)

	return request
}
