package evaluation

import (
	"chess-engine/board"
	"chess-engine/move"
	"chess-engine/util"
)

var pieceValues = map[board.Piece]int{
	board.Pawn:   100,
	board.Knight: 320,
	board.Bishop: 330,
	board.Rook:   500,
	board.Queen:  900,
	board.King:   20000,
}

// Бонусы за контроль центра для пешек и легких фигур
var centerBonus = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 5, 10, 10, 5, 0, 0},
	{0, 5, 10, 20, 20, 10, 5, 0},
	{0, 5, 10, 20, 20, 10, 5, 0},
	{0, 0, 5, 10, 10, 5, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

func Evaluate(b board.Board) int {
	score := 0

	// Битовые маски для подсчета фигур
	var whitePieces, blackPieces uint64
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece, color, _ := b.GetPiece(i, j)
			if piece == board.Empty {
				continue
			}
			// Индекс клетки в 64-битной маске: i*8 + j
			bitPos := uint(i*8 + j)
			if color == board.White {
				whitePieces |= (1 << bitPos)
				score += pieceValues[piece]
				// Бонус за центр для пешек и легких фигур
				if piece == board.Pawn || piece == board.Knight || piece == board.Bishop {
					score += centerBonus[i][j]
				}
			} else {
				blackPieces |= (1 << bitPos)
				score -= pieceValues[piece]
				// Штраф за центр для черных (отзеркаливаем доску)
				if piece == board.Pawn || piece == board.Knight || piece == board.Bishop {
					score -= centerBonus[7-i][j]
				}
			}
		}
	}

	// Подсчет активных фигур
	whiteCount := util.PopCount(whitePieces)
	blackCount := util.PopCount(blackPieces)

	// Бонус за мобильность
	score += (whiteCount - blackCount) * 10

	// Штраф за короля под шахом
	if move.IsKingInCheck(b, board.White) {
		score -= 50
	}
	if move.IsKingInCheck(b, board.Black) {
		score += 50
	}

	return score
}
