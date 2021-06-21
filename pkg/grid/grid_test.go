package grid

import (
	"battlesnake/pkg/types"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloodFill(t *testing.T) {
	tl := []struct {
		name        string
		requestFile string
		point       types.Point
		limit       int
		expected    int
	}{
		{
			name:        "invalid",
			requestFile: "input-4bf799fe-98d5-4cda-a1f9-f3c7cae759eb-turn-92.json",
			point:       types.Point{X: 8, Y: 3},
			limit:       20,
			expected:    0,
		},
		{
			name:        "small",
			requestFile: "input-4bf799fe-98d5-4cda-a1f9-f3c7cae759eb-turn-92.json",
			point:       types.Point{X: 9, Y: 2},
			limit:       20,
			expected:    66,
		},
	}
	for _, tc := range tl {
		t.Run(tc.name, func(t *testing.T) {
			request := parseGameRequest(t, filepath.Join("../../testdata", tc.requestFile))
			grd := Make(request.Board)

			actual := FloodFill(&grd, tc.point, tc.limit)
			require.Equal(t, tc.expected, actual)
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
