package evaluation

import (
	"chess-engine/board"
	"chess-engine/move"
	"chess-engine/util"
	"math"
)

var PieceValues = map[board.Piece]int{
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
				score += PieceValues[piece]
				// Бонус за центр для пешек и легких фигур
				if piece == board.Pawn || piece == board.Knight || piece == board.Bishop {
					score += centerBonus[i][j]
				}
			} else {
				blackPieces |= (1 << bitPos)
				score -= PieceValues[piece]
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

	// Безопасность короля
	score += kingSafety(b, board.White)
	score -= kingSafety(b, board.Black)

	return score
}

// kingSafety оценивает безопасность короля
func kingSafety(b board.Board, color board.Color) int {
	safetyScore := 0

	// Находим позицию короля
	var kingX, kingY int
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			piece, pieceColor, _ := b.GetPiece(x, y)
			if piece == board.King && pieceColor == color {
				kingX, kingY = x, y
				break
			}
		}
	}

	// Проверяем близость фигур противника
	opponentColor := board.Black
	if color == board.Black {
		opponentColor = board.White
	}

	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			piece, pieceColor, _ := b.GetPiece(x, y)
			if piece != board.Empty && pieceColor == opponentColor {
				// Вычисляем расстояние до короля
				distance := int(math.Sqrt(float64((x-kingX)*(x-kingX) + (y-kingY)*(y-kingY))))
				if distance > 0 && distance <= 3 { // Учитываем только близкие фигуры
					switch piece {
					case board.Pawn:
						safetyScore -= 5 / distance
					case board.Knight:
						safetyScore -= 10 / distance
					case board.Bishop:
						safetyScore -= 15 / distance
					case board.Rook:
						safetyScore -= 20 / distance
					case board.Queen:
						safetyScore -= 30 / distance
					}
				}
			}
		}
	}

	// Бонус за пешки рядом с королём (защита)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			nx, ny := kingX+dx, kingY+dy
			if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
				piece, pieceColor, _ := b.GetPiece(nx, ny)
				if piece == board.Pawn && pieceColor == color {
					safetyScore += 10
				}
			}
		}
	}

	return safetyScore
}
