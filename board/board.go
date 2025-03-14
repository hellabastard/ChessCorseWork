package board

import (
	"errors"
)

type Board [8][8]Square

func NewBoard() Board {
	var b Board

	// Расстановка белых фигур
	b[0] = [8]Square{
		// {King, White}, {King, White}, {King, White}, {King, White},
		// {King, White}, {King, White}, {King, White}, {King, White},
		{Rook, White}, {Knight, White}, {Bishop, White}, {Queen, White},
		{King, White}, {Bishop, White}, {Knight, White}, {Rook, White},
	}
	for i := 0; i < 8; i++ {
		b[1][i] = Square{Pawn, White}
	}

	// Расстановка черных фигур
	b[7] = [8]Square{
		// {King, Black}, {King, Black}, {King, Black}, {King, Black},
		// {King, Black}, {King, Black}, {King, Black}, {King, Black},
		{Rook, Black}, {Knight, Black}, {Bishop, Black}, {Queen, Black},
		{King, Black}, {Bishop, Black}, {Knight, Black}, {Rook, Black},
		// {Empty, Black}, {Empty, Black},
	}
	for i := 0; i < 8; i++ {
		b[6][i] = Square{Pawn, Black}
		// b[6][i] = Square{Empty, Black}
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
