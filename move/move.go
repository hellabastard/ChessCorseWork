package move

import (
	"chess-engine/board"
	"errors"
)

type Move struct {
	FromX, FromY int
	ToX, ToY     int
	PromoteTo    board.Piece // Фигура, в которую превращается пешка (0 если нет превращения)
}

// MakeMove выполняет ход на доске
func MakeMove(b *board.Board, m Move) error {
	if m.FromX < 0 || m.FromX >= 8 || m.FromY < 0 || m.FromY >= 8 ||
		m.ToX < 0 || m.ToX >= 8 || m.ToY < 0 || m.ToY >= 8 {
		return errors.New("некорректные координаты хода")
	}

	piece, color, _ := b.GetPiece(m.FromX, m.FromY)
	if piece == board.Empty {
		return errors.New("на начальной клетке нет фигуры")
	}

	// Создаем копию доски и проверяем, не приводит ли ход к шаху
	newBoard := *b
	newBoard.SetPiece(m.ToX, m.ToY, piece, color)
	newBoard.SetPiece(m.FromX, m.FromY, board.Empty, color)

	// Превращение пешки в ферзя
	if piece == board.Pawn {
		if color == board.White && m.ToX == 7 { // Белая пешка на 8-й горизонтали
			if m.PromoteTo != 0 {
				newBoard.SetPiece(m.ToX, m.ToY, m.PromoteTo, color)
			} else {
				newBoard.SetPiece(m.ToX, m.ToY, board.Queen, color) // По умолчанию ферзь
			}
		} else if color == board.Black && m.ToX == 0 { // Чёрная пешка на 1-й горизонтали
			if m.PromoteTo != 0 {
				newBoard.SetPiece(m.ToX, m.ToY, m.PromoteTo, color)
			} else {
				newBoard.SetPiece(m.ToX, m.ToY, board.Queen, color) // По умолчанию ферзь
			}
		}
	}

	// Если это рокировка, перемещаем ладью
	if piece == board.King && abs(m.FromY-m.ToY) == 2 {
		if m.ToY > m.FromY {
			// Короткая рокировка (O-O)
			newBoard.SetPiece(m.FromX, m.FromY+1, board.Rook, color)
			newBoard.SetPiece(m.FromX, m.FromY+3, board.Empty, color)
		} else {
			// Длинная рокировка (O-O-O)
			newBoard.SetPiece(m.FromX, m.FromY-1, board.Rook, color)
			newBoard.SetPiece(m.FromX, m.FromY-4, board.Empty, color)
		}
	}

	// Проверяем, не приводит ли ход к шаху
	if IsKingInCheck(newBoard, color) {
		return errors.New("ход подвергает короля шаху")
	}

	// Если все в порядке, применяем ход
	*b = newBoard
	return nil
}

// IsKingInCheck проверяет, находится ли король под шахом
func IsKingInCheck(b board.Board, color board.Color) bool {
	// Находим позицию короля
	var kingX, kingY int
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece, pieceColor, _ := b.GetPiece(i, j)
			if piece == board.King && pieceColor == color {
				kingX, kingY = i, j
				break
			}
		}
	}

	// Определяем цвет противника
	opponentColor := board.White
	if color == board.White {
		opponentColor = board.Black
	}

	// Проверяем все фигуры противника
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece, pieceColor, _ := b.GetPiece(i, j)
			if pieceColor != opponentColor || piece == board.Empty {
				continue
			}

			// Генерируем возможные ходы фигуры противника
			moves := GenerateMovesForPiece(b, i, j, opponentColor, piece)
			for _, m := range moves {
				if m.ToX == kingX && m.ToY == kingY {
					return true
				}
			}
		}
	}

	return false
}

// Вспомогательная функция для вычисления абсолютного значения
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
