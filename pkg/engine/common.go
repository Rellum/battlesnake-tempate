package engine

import (
	"battlesnake/pkg/types"
	"log"

	"github.com/mitchellh/copystructure"
)

func MoveSnakes(b types.BoardState, moves []types.SnakeMove) types.BoardState {
	r, err := copystructure.Copy(b)
	if err != nil {
		log.Println(err)
	}
	res := r.(types.BoardState)

	for i := 0; i < len(res.Snakes); i++ {
		snake := res.Snakes[i]

		for _, move := range moves {
			if move.ID == snake.ID {
				var newHead = types.Point{}
				switch move.Move {
				case types.MoveDirDown:
					newHead.X = snake.Body[0].X
					newHead.Y = snake.Body[0].Y - 1
				case types.MoveDirLeft:
					newHead.X = snake.Body[0].X - 1
					newHead.Y = snake.Body[0].Y
				case types.MoveDirRight:
					newHead.X = snake.Body[0].X + 1
					newHead.Y = snake.Body[0].Y
				default:
					newHead.X = snake.Body[0].X
					newHead.Y = snake.Body[0].Y + 1
				}

				// Append new head, pop old tail
				snake.Shout = move.Shout
				snake.Health -= 1
				var fi int
				for fi < len(res.Food) {
					if res.Food[fi] != newHead {
						fi++
						continue
					}

					res.Food = append(res.Food[:fi], res.Food[fi+1:]...)
					snake.Health = 100
					fi++
				}
				snake.Body = append([]types.Point{newHead}, snake.Body...)
				if snake.Health != 100 {
					snake.Body = append([]types.Point{}, snake.Body[:len(snake.Body)-1]...)
				}
			}
		}
	}
	return res
}
