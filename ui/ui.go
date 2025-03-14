package ui

import (
	"chess-engine/board"
	"chess-engine/move"
	"chess-engine/search"
	"fmt"
	"image/color"
	"log"

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
	logText              *widget.Entry
	positions            map[string]int // История позиций для правила трёхкратного повторения
}

func NewChessApp() *ChessApp {
	app := &ChessApp{
		currentBoard: board.NewBoard(),
		selectedX:    -1,
		selectedY:    -1,
		logText:      widget.NewEntry(),
		positions:    make(map[string]int), // Инициализируем карту позиций
	}
	// Добавляем начальную позицию в историю
	app.positions[boardToString(app.currentBoard)] = 1
	return app
}

func (appl *ChessApp) Run() {
	myApp := app.New()
	appl.window = myApp.NewWindow("Шахматы")

	appl.grid = appl.createBoardGrid()
	appl.infoLabel = widget.NewLabel("Ваш ход. Выберите фигуру.")
	appl.logText.Disable()
	appl.logText.MultiLine = true

	content := container.NewBorder(
		nil,
		container.NewVBox(appl.infoLabel, appl.logText),
		nil,
		nil,
		appl.grid,
	)
	appl.window.SetContent(content)
	appl.window.Resize(fyne.NewSize(cellSize*8, cellSize*8+100))
	appl.window.Show()
	myApp.Run()
}

func (app *ChessApp) logMessage(msg string) {
	log.Println(msg)
	app.logText.SetText(app.logText.Text + msg + "\n")
}

// boardToString создаёт строковый хэш доски для отслеживания позиций
func boardToString(b board.Board) string {
	var s string
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece, color, _ := b.GetPiece(i, j)
			s += fmt.Sprintf("%d%d", piece, color)
		}
	}
	return s
}

func (app *ChessApp) handleCellClick(x, y int) {
	if app.selectedX == -1 {
		piece, color, err := app.currentBoard.GetPiece(x, y)
		if err != nil {
			app.logMessage(fmt.Sprintf("Ошибка при получении фигуры: %v", err))
			return
		}
		if piece != board.Empty && color == board.White {
			app.selectedX, app.selectedY = x, y
			app.infoLabel.SetText(fmt.Sprintf("Выбрана фигура на %c%d", 'a'+y, x+1))
			app.updateBoard()
		}
	} else {
		piece, color, err := app.currentBoard.GetPiece(x, y)
		if err != nil {
			app.logMessage(fmt.Sprintf("Ошибка при получении фигуры: %v", err))
			return
		}
		if piece != board.Empty && color == board.White {
			app.infoLabel.SetText("Невозможно ходить в клетку с фигурой того же цвета!")
			return
		}

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

		m := move.Move{FromX: app.selectedX, FromY: app.selectedY, ToX: x, ToY: y}
		if err := move.MakeMove(&app.currentBoard, m); err != nil {
			app.infoLabel.SetText("Некорректный ход: " + err.Error())
		} else {
			app.logMessage(fmt.Sprintf("Ход игрока (белые): %s%d-%s%d", string('a'+app.selectedY), app.selectedX+1, string('a'+y), x+1))
			app.selectedX, app.selectedY = -1, -1
			// Обновляем историю позиций
			positionHash := boardToString(app.currentBoard)
			app.positions[positionHash]++
			app.updateBoard()

			if move.IsKingInCheck(app.currentBoard, board.Black) && app.isCheckmate(board.Black) {
				app.infoLabel.SetText("Мат! Белые победили.")
				app.logMessage("Игра завершена: мат черным. Победитель: Белые")
				return
			} else if app.isCheckmate(board.Black) {
				app.infoLabel.SetText("Пат! Ничья.")
				app.logMessage("Игра завершена: пат")
				return
			} else if move.IsKingInCheck(app.currentBoard, board.White) && app.isCheckmate(board.White) {
				app.infoLabel.SetText("Мат! Чёрные победили.")
				app.logMessage("Игра завершена: мат белым. Победитель: Чёрные")
				return
			} else if app.positions[positionHash] >= 3 {
				app.infoLabel.SetText("Ничья по правилу трёхкратного повторения!")
				app.logMessage("Игра завершена: ничья по правилу трёхкратного повторения")
				return
			}
			// Проверка мата или пата для чёрных
			if move.IsKingInCheck(app.currentBoard, board.White) && app.isCheckmate(board.White) {
				app.infoLabel.SetText("Мат! Чёрные победили.")
				app.logMessage("Игра завершена: мат белым. Победитель: Чёрные")
				return
			}

			app.makeAIMove()
		}
	}
}

func (app *ChessApp) makeAIMove() {
	bestMove := search.FindBestMove(app.currentBoard, 3, board.Black)
	if err := move.MakeMove(&app.currentBoard, bestMove); err != nil {
		app.logMessage(fmt.Sprintf("Ошибка при выполнении хода ИИ: %v", err))
		app.infoLabel.SetText("Ошибка ИИ: " + err.Error())
	} else {
		app.logMessage(fmt.Sprintf("Ход ИИ (черные): %s%d-%s%d", string('a'+bestMove.FromY), bestMove.FromX+1, string('a'+bestMove.ToY), bestMove.ToX+1))
		// Обновляем историю позиций
		positionHash := boardToString(app.currentBoard)
		app.positions[positionHash]++
		app.infoLabel.SetText("ИИ сделал ход. Ваш ход.")
		app.updateBoard()

		if move.IsKingInCheck(app.currentBoard, board.White) && app.isCheckmate(board.White) {
			app.infoLabel.SetText("Мат! Черные победили.")
			app.logMessage("Игра завершена: мат белым. Победитель: Черные")
		} else if app.isCheckmate(board.White) {
			app.infoLabel.SetText("Пат! Ничья.")
			app.logMessage("Игра завершена: пат")
		} else if app.positions[positionHash] >= 3 {
			app.infoLabel.SetText("Ничья по правилу трёхкратного повторения!")
			app.logMessage("Игра завершена: ничья по правилу трёхкратного повторения")
		}
	}
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

func (app *ChessApp) isCheckmate(color board.Color) bool {
	moves := move.GenerateMoves(app.currentBoard, color)
	if len(moves) == 0 {
		if move.IsKingInCheck(app.currentBoard, color) {
			return true // Мат
		}
		return false // Пат (ничья, но пока не завершаем игру)
	}
	return false
}

func (app *ChessApp) updateBoard() {
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
