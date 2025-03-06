package move

import (
	"chess-engine/board"
	"errors"
)

// Move представляет ход на шахматной доске
type Move struct {
	FromX, FromY int // Начальная позиция
	ToX, ToY     int // Конечная позиция
}

// GenerateMoves генерирует все возможные ходы для указанного цвета
func GenerateMoves(b board.Board, color board.Color) []Move {
	var moves []Move

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece, pieceColor, _ := b.GetPiece(i, j)
			if pieceColor != color || piece == board.Empty {
				continue // Пропускаем пустые клетки и фигуры другого цвета
			}

			// Генерация ходов в зависимости от типа фигуры
			switch piece {
			case board.Pawn:
				moves = append(moves, generatePawnMoves(b, i, j, color)...)
			case board.Knight:
				moves = append(moves, generateKnightMoves(b, i, j, color)...)
			case board.Bishop:
				moves = append(moves, generateBishopMoves(b, i, j, color)...)
			case board.Rook:
				moves = append(moves, generateRookMoves(b, i, j, color)...)
			case board.Queen:
				moves = append(moves, generateQueenMoves(b, i, j, color)...)
			case board.King:
				moves = append(moves, generateKingMoves(b, i, j, color)...)
			}
		}
	}

	return moves
}

// generatePawnMoves генерирует ходы для пешки
func generatePawnMoves(b board.Board, x, y int, color board.Color) []Move {
	var moves []Move

	direction := 1 // Направление движения пешки (1 для белых, -1 для черных)
	if color == board.Black {
		direction = -1
	}

	// Ход на одну клетку вперед
	if b.IsEmpty(x+direction, y) {
		moves = append(moves, Move{FromX: x, FromY: y, ToX: x + direction, ToY: y})
	}

	// Ход на две клетки вперед (только из начальной позиции)
	if (color == board.White && x == 1) || (color == board.Black && x == 6) {
		if b.IsEmpty(x+direction, y) && b.IsEmpty(x+2*direction, y) {
			moves = append(moves, Move{FromX: x, FromY: y, ToX: x + 2*direction, ToY: y})
		}
	}

	// Взятие фигур по диагонали
	for _, dy := range []int{-1, 1} {
		if !b.IsEmpty(x+direction, y+dy) {
			_, targetColor, _ := b.GetPiece(x+direction, y+dy)
			if targetColor != color {
				moves = append(moves, Move{FromX: x, FromY: y, ToX: x + direction, ToY: y + dy})
			}
		}
	}

	return moves
}

// generateKnightMoves генерирует ходы для коня
func generateKnightMoves(b board.Board, x, y int, color board.Color) []Move {
	var moves []Move

	// Все возможные ходы коня
	knightMoves := [][]int{
		{x + 2, y + 1}, {x + 2, y - 1},
		{x - 2, y + 1}, {x - 2, y - 1},
		{x + 1, y + 2}, {x + 1, y - 2},
		{x - 1, y + 2}, {x - 1, y - 2},
	}

	for _, move := range knightMoves {
		nx, ny := move[0], move[1]
		if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			if b.IsEmpty(nx, ny) {
				moves = append(moves, Move{FromX: x, FromY: y, ToX: nx, ToY: ny})
			} else {
				_, targetColor, _ := b.GetPiece(nx, ny)
				if targetColor != color {
					moves = append(moves, Move{FromX: x, FromY: y, ToX: nx, ToY: ny})
				}
			}
		}
	}

	return moves
}

// generateBishopMoves генерирует ходы для слона
func generateBishopMoves(b board.Board, x, y int, color board.Color) []Move {
	return generateDiagonalMoves(b, x, y, color)
}

// generateRookMoves генерирует ходы для ладьи
func generateRookMoves(b board.Board, x, y int, color board.Color) []Move {
	return generateStraightMoves(b, x, y, color)
}

// generateQueenMoves генерирует ходы для ферзя
func generateQueenMoves(b board.Board, x, y int, color board.Color) []Move {
	// Ферзь сочетает возможности ладьи и слона
	moves := generateStraightMoves(b, x, y, color)
	moves = append(moves, generateDiagonalMoves(b, x, y, color)...)
	return moves
}

// generateKingMoves генерирует ходы для короля
func generateKingMoves(b board.Board, x, y int, color board.Color) []Move {
	var moves []Move

	// Все возможные ходы короля
	kingMoves := [][]int{
		{x + 1, y}, {x - 1, y},
		{x, y + 1}, {x, y - 1},
		{x + 1, y + 1}, {x + 1, y - 1},
		{x - 1, y + 1}, {x - 1, y - 1},
	}

	for _, move := range kingMoves {
		nx, ny := move[0], move[1]
		if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			if b.IsEmpty(nx, ny) {
				moves = append(moves, Move{FromX: x, FromY: y, ToX: nx, ToY: ny})
			} else {
				_, targetColor, _ := b.GetPiece(nx, ny)
				if targetColor != color {
					moves = append(moves, Move{FromX: x, FromY: y, ToX: nx, ToY: ny})
				}
			}
		}
	}

	return moves
}

// generateDiagonalMoves генерирует ходы по диагонали (для слона и ферзя)
func generateDiagonalMoves(b board.Board, x, y int, color board.Color) []Move {
	var moves []Move

	directions := [][]int{
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}

	for _, dir := range directions {
		dx, dy := dir[0], dir[1]
		nx, ny := x+dx, y+dy
		for nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			if b.IsEmpty(nx, ny) {
				moves = append(moves, Move{FromX: x, FromY: y, ToX: nx, ToY: ny})
			} else {
				_, targetColor, _ := b.GetPiece(nx, ny)
				if targetColor != color {
					moves = append(moves, Move{FromX: x, FromY: y, ToX: nx, ToY: ny})
				}
				break // Прерываем цикл, если нашли фигуру
			}
			nx += dx
			ny += dy
		}
	}

	return moves
}

// generateStraightMoves генерирует ходы по прямой (для ладьи и ферзя)
func generateStraightMoves(b board.Board, x, y int, color board.Color) []Move {
	var moves []Move

	directions := [][]int{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
	}

	for _, dir := range directions {
		dx, dy := dir[0], dir[1]
		nx, ny := x+dx, y+dy
		for nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			if b.IsEmpty(nx, ny) {
				moves = append(moves, Move{FromX: x, FromY: y, ToX: nx, ToY: ny})
			} else {
				_, targetColor, _ := b.GetPiece(nx, ny)
				if targetColor != color {
					moves = append(moves, Move{FromX: x, FromY: y, ToX: nx, ToY: ny})
				}
				break // Прерываем цикл, если нашли фигуру
			}
			nx += dx
			ny += dy
		}
	}

	return moves
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

	// Выполняем ход
	b.SetPiece(m.ToX, m.ToY, piece, color)
	b.SetPiece(m.FromX, m.FromY, board.Empty, color)

	return nil
}
