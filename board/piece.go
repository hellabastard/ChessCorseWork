package board

type Piece int

const (
	Empty Piece = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

type Color int

const (
	White Color = iota
	Black
)

type Square struct {
	Piece Piece
	Color Color
}
