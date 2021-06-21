package types

import (
	"bytes"
	"fmt"

	"github.com/BattlesnakeOfficial/rules"
)

var hChars = []rune{'■', '●', '◘', '☻'}
var chars = []rune{'□', '⍟', '⌀', '☺'}

func Print(state BoardState, outOfBounds []rules.Point) []byte {
	var o bytes.Buffer
	board := make([][]rune, state.Width)
	for i := range board {
		board[i] = make([]rune, state.Height)
	}
	for y := 0; y < state.Height; y++ {
		for x := 0; x < state.Width; x++ {
			board[x][y] = '◦'
		}
	}
	for _, oob := range outOfBounds {
		board[oob.X][oob.Y] = '░'
	}
	if len(outOfBounds) > 0 {
		o.WriteString(fmt.Sprintf("Hazards ░: %v\n", outOfBounds))
	}
	for _, f := range state.Food {
		board[f.X][f.Y] = '⚕'
	}
	o.WriteString(fmt.Sprintf("Food ⚕: %v\n", state.Food))
	for i, s := range state.Snakes {
		for j, b := range s.Body {
			if j == 0 {
				board[b.X][b.Y] = hChars[i]
			} else {
				board[b.X][b.Y] = chars[i]
			}
		}
		o.WriteString(fmt.Sprintf("%v %c%c%c: %v\n", s.Name, hChars[i], chars[i], chars[i], s))
	}
	for y := state.Height - 1; y >= 0; y-- {
		for x := 0; x < state.Width; x++ {
			o.WriteRune(board[x][y])
		}
		o.WriteString("\n")
	}
	return o.Bytes()
}
