package striker

import (
	"battlesnake/pkg/types"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMove(t *testing.T) {
	tests := []struct {
		requestFile string
		possible    []types.MoveDir
	}{
		{
			requestFile: "case_1.json",
			possible:    []types.MoveDir{"up", "down", "left", "right"},
		},
		{
			requestFile: "case_2.json",
			possible:    []types.MoveDir{"left", "right"},
		},
		{
			requestFile: "case_3.json",
			possible:    []types.MoveDir{"up", "right"},
		},
		{
			requestFile: "case_4.json",
			possible:    []types.MoveDir{"down"},
		},
		{
			requestFile: "case_5.json",
			possible:    []types.MoveDir{"left", "right"},
		},
		{
			requestFile: "case_6.json",
			possible:    []types.MoveDir{"left", "right"},
		},
		{
			requestFile: "case_7.json",
			possible:    []types.MoveDir{"left", "right"},
		},
		{
			requestFile: "case_8.json",
			possible:    []types.MoveDir{"right"},
		},
		{
			requestFile: "case_9.json",
			possible:    []types.MoveDir{"up"},
		},
		{
			requestFile: "case_10.json",
			possible:    []types.MoveDir{"left", "right"},
		},
		{
			requestFile: "case_11.json",
			possible:    []types.MoveDir{"down"},
		},
		{
			requestFile: "case_12.json",
			possible:    []types.MoveDir{"left", "down"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.requestFile, func(t *testing.T) {
			file, err := os.Open(path.Join("testdata", tc.requestFile))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			HandleMove(recorder, httptest.NewRequest(http.MethodPost, "/move", file))

			var response types.SnakeMove
			err = json.NewDecoder(recorder.Body).Decode(&response)
			require.NoError(t, err)

			require.Contains(t, tc.possible, response.Move)
		})
	}
}
