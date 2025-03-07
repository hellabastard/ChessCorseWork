package evaluation

import "chess-engine/board"

var pieceValues = map[board.Piece]int{
	board.Pawn:   100,
	board.Knight: 320,
	board.Bishop: 330,
	board.Rook:   500,
	board.Queen:  900,
	board.King:   20000,
}

func Evaluate(b board.Board) int {
	score := 0

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece, color, _ := b.GetPiece(i, j)
			if piece == board.Empty {
				continue
			}

			if color == board.White {
				score += pieceValues[piece]
			} else {
				score -= pieceValues[piece]
			}
		}
	}

	score += evaluatePositionalFactors(b)

	return score
}

func evaluatePositionalFactors(b board.Board) int {
	positionalScore := 0

	centerSquares := [][]int{
		{3, 3}, {3, 4}, {4, 3}, {4, 4},
	}

	for _, square := range centerSquares {
		x, y := square[0], square[1]
		piece, color, _ := b.GetPiece(x, y)

		if piece != board.Empty {
			if color == board.White {
				positionalScore += 10
			} else {
				positionalScore -= 10
			}
		}
	}

	return positionalScore
}
