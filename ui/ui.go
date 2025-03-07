package ui

import (
	"chess-engine/board"
	"chess-engine/move"
	"chess-engine/search"
	"fmt"
	"image/color"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	cellSize = 80
)

var (
	pieceColors = map[board.Color]color.Color{
		board.White: color.White,
		board.Black: color.Black,
	}
	highlightColor     = color.RGBA{R: 0, G: 255, B: 0, A: 50}
	selectedColor      = color.RGBA{R: 255, G: 255, B: 0, A: 50}
	availableMoveColor = color.RGBA{R: 0, G: 255, B: 0, A: 50}
)

type ChessApp struct {
	currentBoard         board.Board
	selectedX, selectedY int
	window               fyne.Window
	grid                 *fyne.Container
	infoLabel            *widget.Label
}

func NewChessApp() *ChessApp {
	return &ChessApp{
		currentBoard: board.NewBoard(),
		selectedX:    -1,
		selectedY:    -1,
	}
}

func (appl *ChessApp) Run() {
	myApp := app.New()
	appl.window = myApp.NewWindow("Шахматы")

	appl.grid = appl.createBoardGrid()
	appl.infoLabel = widget.NewLabel("Ваш ход. Выберите фигуру.")

	content := container.NewBorder(nil, appl.infoLabel, nil, nil, appl.grid)
	appl.window.SetContent(content)
	appl.window.Resize(fyne.NewSize(cellSize*8, cellSize*8+50))
	appl.window.Show()
	myApp.Run()
}

func (app *ChessApp) createCell(x, y int) fyne.CanvasObject {
	lightColor := color.RGBA{R: 240, G: 217, B: 181, A: 255}
	darkColor := color.RGBA{R: 181, G: 136, B: 99, A: 255}

	cellColor := lightColor
	if (x+y)%2 == 1 {
		cellColor = darkColor
	}

	background := canvas.NewRectangle(cellColor)
	background.SetMinSize(fyne.NewSize(cellSize, cellSize))

	var figure fyne.CanvasObject
	piece, pieceColor, err := app.currentBoard.GetPiece(x, y)
	if err != nil {
		log.Printf("Ошибка при получении фигуры: %v", err)
	}
	if piece != board.Empty {
		figure = app.createFigure(piece, pieceColor)
	} else {
		figure = canvas.NewRectangle(color.Transparent)
	}

	cellContainer := container.NewMax(
		background,
		container.NewCenter(figure),
	)

	if x == app.selectedX && y == app.selectedY {
		highlight := canvas.NewRectangle(selectedColor)
		highlight.SetMinSize(fyne.NewSize(cellSize, cellSize))
		cellContainer.Add(highlight)
	}

	if app.selectedX != -1 && app.selectedY != -1 {
		availableMoves := app.getAvailableMoves(app.selectedX, app.selectedY)
		for _, m := range availableMoves {
			if m.ToX == x && m.ToY == y {
				availableHighlight := canvas.NewRectangle(availableMoveColor)
				availableHighlight.SetMinSize(fyne.NewSize(cellSize, cellSize))
				cellContainer.Add(availableHighlight)
			}
		}
	}

	button := NewCustomButton("", func() {
		app.handleCellClick(x, y)
	}, func() {
		app.handleRightClick()
	})
	button.Importance = widget.LowImportance
	button.Resize(fyne.NewSize(cellSize, cellSize))

	cellContainer.Add(button)

	return cellContainer
}

func (app *ChessApp) createFigure(piece board.Piece, color board.Color) fyne.CanvasObject {
	var figure fyne.CanvasObject

	figureColor := pieceColors[color]

	switch piece {
	case board.Pawn:
		figure = canvas.NewText("♙", figureColor)
	case board.Knight:
		figure = canvas.NewText("♘", figureColor)
	case board.Bishop:
		figure = canvas.NewText("♗", figureColor)
	case board.Rook:
		figure = canvas.NewText("♖", figureColor)
	case board.Queen:
		figure = canvas.NewText("♕", figureColor)
	case board.King:
		figure = canvas.NewText("♔", figureColor)
	default:
		figure = canvas.NewText("", figureColor)
	}

	figure.(*canvas.Text).TextSize = cellSize / 2
	figure.(*canvas.Text).Alignment = fyne.TextAlignCenter
	figure.Resize(fyne.NewSize(cellSize, cellSize))

	return figure
}

func (app *ChessApp) handleCellClick(x, y int) {
	if app.selectedX == -1 {
		// Выбор фигуры (игрок может ходить только белыми)
		piece, color, err := app.currentBoard.GetPiece(x, y)
		if err != nil {
			log.Printf("Ошибка при получении фигуры: %v", err)
			return
		}
		if piece != board.Empty && color == board.White {
			app.selectedX, app.selectedY = x, y
			app.infoLabel.SetText(fmt.Sprintf("Выбрана фигура на %c%d", 'a'+y, x+1))
			app.updateBoard()
		}
	} else {
		// Проверка, можно ли ходить в выбранную клетку
		piece, color, err := app.currentBoard.GetPiece(x, y)
		if err != nil {
			log.Printf("Ошибка при получении фигуры: %v", err)
			return
		}

		// Если в клетке находится фигура того же цвета, ход невозможен
		if piece != board.Empty && color == board.White {
			app.infoLabel.SetText("Невозможно ходить в клетку с фигурой того же цвета!")
			return
		}

		// Проверяем, является ли клетка допустимой для хода
		availableMoves := app.getAvailableMoves(app.selectedX, app.selectedY)
		isValidMove := false
		for _, m := range availableMoves {
			if m.ToX == x && m.ToY == y {
				isValidMove = true
				break
			}
		}

		if !isValidMove {
			app.infoLabel.SetText("Невозможно ходить в эту клетку!")
			return
		}

		// Выполнение хода игрока (белыми)
		m := move.Move{FromX: app.selectedX, FromY: app.selectedY, ToX: x, ToY: y}
		if err := move.MakeMove(&app.currentBoard, m); err != nil {
			app.infoLabel.SetText("Некорректный ход: " + err.Error())
		} else {
			app.selectedX, app.selectedY = -1, -1
			app.updateBoard()

			// Проверка на мат
			if app.isCheckmate(board.Black) {
				app.infoLabel.SetText("Мат! Игра окончена.")
				os.Exit(0)
			}

			// Переход хода к черным (ИИ)
			app.makeAIMove()
		}
	}
}

// func (app *ChessApp) handleCellClick(x, y int) {
// 	if app.selectedX == -1 {
// 		// Выбор фигуры (игрок может ходить только белыми)
// 		piece, color, err := app.currentBoard.GetPiece(x, y)
// 		if err != nil {
// 			log.Printf("Ошибка при получении фигуры: %v", err)
// 			return
// 		}
// 		if piece != board.Empty && color == board.White {
// 			app.selectedX, app.selectedY = x, y
// 			app.infoLabel.SetText(fmt.Sprintf("Выбрана фигура на %c%d", 'a'+y, x+1))
// 			app.updateBoard()
// 		}
// 	} else {
// 		// Проверка, можно ли ходить в выбранную клетку
// 		piece, color, err := app.currentBoard.GetPiece(x, y)
// 		if err != nil {
// 			log.Printf("Ошибка при получении фигуры: %v", err)
// 			return
// 		}

// 		// Если в клетке находится фигура того же цвета, ход невозможен
// 		if piece != board.Empty && color == board.White {
// 			app.infoLabel.SetText("Невозможно ходить в клетку с фигурой того же цвета!")
// 			return
// 		}

// 		// Проверяем, является ли клетка допустимой для хода
// 		availableMoves := app.getAvailableMoves(app.selectedX, app.selectedY)
// 		isValidMove := false
// 		for _, m := range availableMoves {
// 			if m.ToX == x && m.ToY == y {
// 				isValidMove = true
// 				break
// 			}
// 		}

// 		if !isValidMove {
// 			app.infoLabel.SetText("Невозможно ходить в эту клетку!")
// 			return
// 		}

// 		// Выполнение хода игрока (белыми)
// 		m := move.Move{FromX: app.selectedX, FromY: app.selectedY, ToX: x, ToY: y}
// 		if err := move.MakeMove(&app.currentBoard, m); err != nil {
// 			app.infoLabel.SetText("Некорректный ход: " + err.Error())
// 		} else {
// 			app.selectedX, app.selectedY = -1, -1
// 			app.updateBoard()

// 			// Проверка на мат
// 			if app.isCheckmate(board.Black) {
// 				app.infoLabel.SetText("Мат! Игра окончена.")
// 				os.Exit(0)
// 			}

// 			// Переход хода к черным (ИИ)
// 			app.makeAIMove()
// 		}
// 	}
// }

func (app *ChessApp) handleRightClick() {
	app.selectedX, app.selectedY = -1, -1
	app.infoLabel.SetText("Выбранная фигура сброшена")
	app.updateBoard()
}

func (app *ChessApp) getAvailableMoves(x, y int) []move.Move {
	piece, color, err := app.currentBoard.GetPiece(x, y)
	if err != nil || piece == board.Empty {
		return nil
	}

	allMoves := move.GenerateMoves(app.currentBoard, color)

	var availableMoves []move.Move
	for _, m := range allMoves {
		if m.FromX == x && m.FromY == y {
			availableMoves = append(availableMoves, m)
		}
	}

	return availableMoves
}

func (app *ChessApp) makeAIMove() {
	log.Println("ИИ делает ход за черных...")
	bestMove := search.FindBestMove(app.currentBoard, 3, board.Black) // ИИ ходит за черных
	if err := move.MakeMove(&app.currentBoard, bestMove); err != nil {
		log.Println("Ошибка ИИ:", err)
	} else {
		app.infoLabel.SetText("ИИ сделал ход. Ваш ход.")
		app.updateBoard()

		// Проверка на мат
		if app.isCheckmate(board.White) {
			app.infoLabel.SetText("Мат! Игра окончена.")
			os.Exit(0)
		}
	}
}

func (app *ChessApp) isCheckmate(color board.Color) bool {
	moves := move.GenerateMoves(app.currentBoard, color)
	return len(moves) == 0
}

func (app *ChessApp) updateBoard() {
	log.Println("Обновление доски...")
	app.grid = app.createBoardGrid()
	app.window.SetContent(container.NewBorder(nil, app.infoLabel, nil, nil, app.grid))
	app.window.Content().Refresh()
}

func (app *ChessApp) createBoardGrid() *fyne.Container {
	grid := container.NewGridWithColumns(8) // Создаем сетку 8x8
	for i := 7; i >= 0; i-- {               // Идем сверху вниз (для корректного отображения доски)
		for j := 0; j < 8; j++ {
			cell := app.createCell(i, j) // Создаем клетку
			grid.Add(cell)               // Добавляем клетку в сетку
		}
	}
	return grid
}
