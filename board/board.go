package board

import (
	"errors"
	"fmt"
)

// Piece представляет тип шахматной фигуры
type Piece int

// Константы для типов фигур
const (
	Empty Piece = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

// Color представляет цвет фигуры
type Color int

// Константы для цветов
const (
	White Color = iota
	Black
)

// Square представляет клетку на шахматной доске
type Square struct {
	Piece Piece
	Color Color
}

// Board представляет шахматную доску
type Board [8][8]Square

// NewBoard создает новую доску с начальной расстановкой фигур
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
			b[i][j] = Square{Empty, White} // Цвет для пустых клеток не имеет значения
		}
	}

	return b
}

// PrintBoard выводит доску в консоль
func (b Board) PrintBoard() {
	for i := 7; i >= 0; i-- {
		for j := 0; j < 8; j++ {
			square := b[i][j]
			var pieceSymbol string
			switch square.Piece {
			case Empty:
				pieceSymbol = "."
			case Pawn:
				pieceSymbol = "P"
			case Knight:
				pieceSymbol = "N"
			case Bishop:
				pieceSymbol = "B"
			case Rook:
				pieceSymbol = "R"
			case Queen:
				pieceSymbol = "Q"
			case King:
				pieceSymbol = "K"
			}
			if square.Color == Black {
				pieceSymbol = string(pieceSymbol[0] + 32) // Преобразуем в нижний регистр для черных фигур
			}
			fmt.Printf("%2s ", pieceSymbol)
		}
		fmt.Println()
	}
}

// GetPiece возвращает фигуру на указанной клетке
func (b Board) GetPiece(x, y int) (Piece, Color, error) {
	if x < 0 || x >= 8 || y < 0 || y >= 8 {
		return Empty, White, errors.New("координаты за пределами доски")
	}
	square := b[x][y]
	return square.Piece, square.Color, nil
}

// SetPiece устанавливает фигуру на указанную клетку
func (b *Board) SetPiece(x, y int, piece Piece, color Color) error {
	if x < 0 || x >= 8 || y < 0 || y >= 8 {
		return errors.New("координаты за пределами доски")
	}
	b[x][y] = Square{piece, color}
	return nil
}

// IsEmpty проверяет, пуста ли клетка
func (b Board) IsEmpty(x, y int) bool {
	if x < 0 || x >= 8 || y < 0 || y >= 8 {
		return false
	}
	return b[x][y].Piece == Empty
}
