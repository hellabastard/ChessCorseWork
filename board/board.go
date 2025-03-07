package board

import (
	"errors"
)

type Board [8][8]Square

func NewBoard() Board {
	var b Board

	// Расстановка белых фигур
	b[0] = [8]Square{
		{Rook, White}, {Knight, White}, {Bishop, White}, {Queen, White},
		{King, White}, {Bishop, White}, {Knight, White}, {Rook, White},
	}
	for i := 0; i < 8; i++ {
		b[1][i] = Square{Pawn, White}
	}

	// Расстановка черных фигур
	b[7] = [8]Square{
		{Rook, Black}, {Knight, Black}, {Bishop, Black}, {Queen, Black},
		{King, Black}, {Bishop, Black}, {Knight, Black}, {Rook, Black},
	}
	for i := 0; i < 8; i++ {
		b[6][i] = Square{Pawn, Black}
	}

	// Остальные клетки пустые
	for i := 2; i < 6; i++ {
		for j := 0; j < 8; j++ {
			b[i][j] = Square{Empty, White}
		}
	}

	return b
}

func (b Board) GetPiece(x, y int) (Piece, Color, error) {
	if x < 0 || x >= 8 || y < 0 || y >= 8 {
		return Empty, White, errors.New("координаты за пределами доски")
	}
	return b[x][y].Piece, b[x][y].Color, nil
}

func (b *Board) SetPiece(x, y int, piece Piece, color Color) error {
	if x < 0 || x >= 8 || y < 0 || y >= 8 {
		return errors.New("координаты за пределами доски")
	}
	b[x][y] = Square{piece, color}
	return nil
}

func (b Board) IsEmpty(x, y int) bool {
	if x < 0 || x >= 8 || y < 0 || y >= 8 {
		return false
	}
	return b[x][y].Piece == Empty
}

// IsKingInCheck проверяет, находится ли король под шахом
func (b Board) IsKingInCheck(color Color) bool {
	// Находим позицию короля
	var kingX, kingY int
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece, pieceColor, _ := b.GetPiece(i, j)
			if piece == King && pieceColor == color {
				kingX, kingY = i, j
				break
			}
		}
	}

	// Определяем цвет противника
	opponentColor := White
	if color == White {
		opponentColor = Black
	}

	// Проверяем, атакуют ли короля фигуры противника
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue // Пропускаем клетку короля
			}
			x, y := kingX+dx, kingY+dy
			if x < 0 || x >= 8 || y < 0 || y >= 8 {
				continue
			}
			piece, pieceColor, _ := b.GetPiece(x, y)
			if pieceColor == opponentColor {
				// Проверяем, может ли фигура атаковать короля
				switch piece {
				case Pawn:
					if (color == White && dx == -1) || (color == Black && dx == 1) {
						if dy == -1 || dy == 1 {
							return true
						}
					}
				case Knight:
					if (abs(dx) == 2 && abs(dy) == 1) || (abs(dx) == 1 && abs(dy) == 2) {
						return true
					}
				case Bishop:
					if abs(dx) == abs(dy) {
						return true
					}
				case Rook:
					if dx == 0 || dy == 0 {
						return true
					}
				case Queen:
					if abs(dx) == abs(dy) || dx == 0 || dy == 0 {
						return true
					}
				case King:
					if abs(dx) <= 1 && abs(dy) <= 1 {
						return true
					}
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
