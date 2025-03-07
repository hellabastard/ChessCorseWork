package move

import (
	"chess-engine/board"
	"errors"
)

type Move struct {
	FromX, FromY int
	ToX, ToY     int
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
	if newBoard.IsKingInCheck(color) {
		return errors.New("ход подвергает короля шаху")
	}

	// Если все в порядке, применяем ход
	*b = newBoard
	return nil
}

// Вспомогательная функция для вычисления абсолютного значения
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
