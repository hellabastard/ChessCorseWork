package main

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
	cellSize = 80 // Размер клетки
)

var (
	pieceColors = map[board.Color]color.Color{
		board.White: color.White,
		board.Black: color.Black,
	}
	highlightColor     = color.RGBA{R: 0, G: 255, B: 0, A: 50}   // Зелёный с прозрачностью
	selectedColor      = color.RGBA{R: 255, G: 255, B: 0, A: 50} // Жёлтый для выбранной клетки
	availableMoveColor = color.RGBA{R: 0, G: 255, B: 0, A: 50}   // Зелёный для доступных ходов
)

// CustomButton - кастомный виджет для обработки правого клика
type CustomButton struct {
	widget.Button
	onRightClick func()
}

func NewCustomButton(label string, onLeftClick, onRightClick func()) *CustomButton {
	b := &CustomButton{
		Button: widget.Button{
			Text:     label,
			OnTapped: onLeftClick,
		},
		onRightClick: onRightClick,
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *CustomButton) TappedSecondary(*fyne.PointEvent) {
	if b.onRightClick != nil {
		b.onRightClick()
	}
}

type ChessApp struct {
	currentBoard         board.Board
	selectedX, selectedY int // Выбранная клетка
	window               fyne.Window
	grid                 *fyne.Container
	infoLabel            *widget.Label
}

func NewChessApp() *ChessApp {
	return &ChessApp{
		currentBoard: board.NewBoard(),
		selectedX:    -1, // -1 означает, что клетка не выбрана
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

func (app *ChessApp) createBoardGrid() *fyne.Container {
	grid := container.NewGridWithColumns(8)
	for i := 7; i >= 0; i-- {
		for j := 0; j < 8; j++ {
			cell := app.createCell(i, j)
			grid.Add(cell)
		}
	}
	return grid
}

func (app *ChessApp) createCell(x, y int) fyne.CanvasObject {
	// Цвета для клеток
	lightColor := color.RGBA{R: 240, G: 217, B: 181, A: 255} // Светлая клетка
	darkColor := color.RGBA{R: 181, G: 136, B: 99, A: 255}   // Тёмная клетка

	// Выбираем цвет клетки
	cellColor := lightColor
	if (x+y)%2 == 1 {
		cellColor = darkColor
	}

	// Создаем фон клетки
	background := canvas.NewRectangle(cellColor)
	background.SetMinSize(fyne.NewSize(cellSize, cellSize))

	// Создаем фигуру
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

	// Создаем контейнер для клетки
	cellContainer := container.NewMax(
		background,
		container.NewCenter(figure),
	)

	// Добавляем выделение, если клетка выбрана
	if x == app.selectedX && y == app.selectedY {
		highlight := canvas.NewRectangle(selectedColor)
		highlight.SetMinSize(fyne.NewSize(cellSize, cellSize))
		cellContainer.Add(highlight)
	}

	// Добавляем подсветку для доступных ходов
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

	// Создаем кастомную кнопку для обработки кликов
	button := NewCustomButton("", func() {
		app.handleCellClick(x, y)
	}, func() {
		app.handleRightClick(x, y)
	})
	button.Importance = widget.LowImportance        // Убираем стили кнопки
	button.Resize(fyne.NewSize(cellSize, cellSize)) // Устанавливаем размер кнопки

	// Добавляем кнопку поверх всего
	cellContainer.Add(button)

	return cellContainer
}

func (app *ChessApp) createFigure(piece board.Piece, color board.Color) fyne.CanvasObject {
	var figure fyne.CanvasObject

	// Определяем цвет фигуры
	figureColor := pieceColors[color]

	// Создаем фигуру в зависимости от её типа
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
		figure = canvas.NewText("", figureColor) // Пустая фигура
	}

	// Настраиваем размер и шрифт фигуры
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
			app.updateBoard() // Обновляем доску для отображения выделения и доступных ходов
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
			app.infoLabel.SetText("Некорректный ход!")
		} else {
			app.selectedX, app.selectedY = -1, -1
			app.updateBoard() // Обновляем доску после хода игрока

			// Проверка на мат
			if app.isCheckmate(board.Black) {
				app.infoLabel.SetText("Мат! Игра окончена.")
				os.Exit(0) // Завершаем программу
			}

			// Переход хода к черным (ИИ)
			app.makeAIMove()
		}
	}
}

func (app *ChessApp) handleRightClick(x, y int) {
	// Сбрасываем выбранную фигуру
	app.selectedX, app.selectedY = -1, -1
	app.updateBoard() // Обновляем доску
}

func (app *ChessApp) getAvailableMoves(x, y int) []move.Move {
	piece, color, err := app.currentBoard.GetPiece(x, y)
	if err != nil || piece == board.Empty {
		return nil
	}

	// Генерация всех возможных ходов для фигуры
	allMoves := move.GenerateMoves(app.currentBoard, color)

	// Фильтруем ходы, чтобы оставить только те, которые начинаются с выбранной клетки
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
		app.updateBoard() // Обновляем доску после хода ИИ

		// Проверка на мат
		if app.isCheckmate(board.White) {
			app.infoLabel.SetText("Мат! Игра окончена.")
			os.Exit(0) // Завершаем программу
		}
	}
}

func (app *ChessApp) isCheckmate(color board.Color) bool {
	// Проверяем, есть ли у игрока допустимые ходы
	moves := move.GenerateMoves(app.currentBoard, color)
	return len(moves) == 0
}

func (app *ChessApp) updateBoard() {
	log.Println("Обновление доски...")
	app.grid = app.createBoardGrid() // Пересоздаем сетку с новым состоянием доски
	app.window.SetContent(container.NewBorder(nil, app.infoLabel, nil, nil, app.grid))
	app.window.Content().Refresh() // Принудительно обновляем содержимое окна
}

func main() {
	chessApp := NewChessApp()
	chessApp.Run()
}
