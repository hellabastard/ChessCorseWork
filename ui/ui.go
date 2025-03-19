package ui

import (
	"chess-engine/board"
	"chess-engine/evaluation"
	"chess-engine/move"
	"chess-engine/search"
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
	cellSize = 80
)

var (
	pieceColors = map[board.Color]color.Color{
		board.White: color.White,
		board.Black: color.Black,
	}
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
	commandEntry         *widget.Entry
	positions            map[string]int
	gameOver             bool
	aiThinking           bool
	moveCount            int
	aiColor              board.Color
	lastStats            search.SearchStats
	moveHistory          []move.Move
}

func NewChessApp() *ChessApp {
	search.LoadData()
	app := &ChessApp{
		currentBoard: board.NewBoard(),
		selectedX:    -1,
		selectedY:    -1,
		logText:      widget.NewEntry(),
		commandEntry: widget.NewEntry(),
		positions:    make(map[string]int),
		gameOver:     false,
		aiThinking:   false,
		moveCount:    0,
		aiColor:      board.Black,
		moveHistory:  make([]move.Move, 0),
	}
	app.positions[boardToString(app.currentBoard)] = 1
	return app
}

func (appl *ChessApp) Run() {
	myApp := app.New()
	appl.window = myApp.NewWindow("Шахматы")

	appl.grid = appl.createBoardGrid()
	appl.infoLabel = widget.NewLabel("Ваш ход. Выберите фигуру или введите команду.")
	appl.logText.MultiLine = true
	appl.logText.Wrapping = fyne.TextWrapWord
	appl.logText.Disable()

	appl.commandEntry.OnSubmitted = func(input string) {
		appl.processCommand(input)
		appl.commandEntry.SetText("")
	}

	logContainer := container.NewMax(
		canvas.NewRectangle(color.RGBA{R: 30, G: 30, B: 30, A: 255}),
		appl.logText,
	)

	content := container.NewBorder(
		nil,
		container.NewVBox(
			appl.infoLabel,
			logContainer,
			container.NewBorder(nil, nil, widget.NewLabel("Команда:"), nil, appl.commandEntry),
		),
		nil,
		nil,
		appl.grid,
	)
	appl.window.SetContent(content)
	appl.window.Resize(fyne.NewSize(cellSize*8, cellSize*8+150))

	appl.window.SetCloseIntercept(func() {
		search.SaveData()
		appl.window.Close()
	})

	appl.window.Show()
	myApp.Run()
}

func (app *ChessApp) processCommand(input string) {
	if app.aiThinking {
		app.infoLabel.SetText("Подождите, ИИ думает...")
		return
	}
	parts := strings.Split(strings.TrimSpace(input), " ")

	switch parts[0] {
	case "play":
		if len(parts) < 2 {
			app.logMessage("Ошибка: укажите ход, например 'play e2e4'")
			return
		}
		moveStr := parts[1]
		if len(moveStr) != 4 {
			app.logMessage("Ошибка: ход должен быть в формате 'e2e4'")
			return
		}
		fromY := int(moveStr[0] - 'a')
		fromX, _ := strconv.Atoi(string(moveStr[1]))
		toY := int(moveStr[2] - 'a')
		toX, _ := strconv.Atoi(string(moveStr[3]))
		fromX, toX = fromX-1, toX-1
		m := move.Move{FromX: fromX, FromY: fromY, ToX: toX, ToY: toY}

		currentColor := app.currentPlayerColor()
		legalMoves := move.GenerateMoves(app.currentBoard, currentColor)
		isLegal := false
		for _, legalMove := range legalMoves {
			if legalMove == m {
				isLegal = true
				break
			}
		}
		if !isLegal {
			app.logMessage("Ошибка: ход не является легальным")
			return
		}

		if err := move.MakeMove(&app.currentBoard, m); err != nil {
			app.logMessage(fmt.Sprintf("Ошибка хода: %v", err))
		} else {
			app.logMessage(fmt.Sprintf("Ход: %s%d-%s%d", string('a'+fromY), fromX+1, string('a'+toY), toX+1))
			app.playMoveSound()
			app.moveCount++
			app.moveHistory = append(app.moveHistory, m)
			app.checkGameEnd(app.aiColor)
			app.updateBoard()
			app.logMessage(fmt.Sprintf("После хода игрока: moveCount=%d, currentColor=%v, aiColor=%v", app.moveCount, app.currentPlayerColor(), app.aiColor))
			// Автоматический ход ИИ
			if !app.gameOver && app.aiColor == oppositeColor(currentColor) {
				app.logMessage("Вызываю makeAIMove для ИИ")
				app.makeAIMove(4) // Убрали go для синхронности
			}
		}

	case "move":
		if len(parts) < 2 {
			app.logMessage("Ошибка: укажите глубину, например 'move 4'")
			return
		}
		depth, err := strconv.Atoi(parts[1])
		if err != nil || depth <= 0 {
			app.logMessage("Ошибка: глубина должна быть положительным числом")
			return
		}
		app.makeAIMove(depth)

	case "print":
		app.logMessage(boardToString(app.currentBoard))

	case "stats":
		if app.lastStats.NodesEvaluated == 0 {
			app.logMessage("Статистика недоступна: выполните поиск хода")
		} else {
			app.logMessage(fmt.Sprintf("Оценено узлов: %d", app.lastStats.NodesEvaluated))
			app.logMessage(fmt.Sprintf("Время поиска: %v", app.lastStats.SearchTime))
			app.logMessage(fmt.Sprintf("Скорость: %.2f узлов/с", float64(app.lastStats.NodesEvaluated)/app.lastStats.SearchTime.Seconds()))
		}

	case "reset":
		app.currentBoard = board.NewBoard()
		app.selectedX, app.selectedY = -1, -1
		app.positions = make(map[string]int)
		app.positions[boardToString(app.currentBoard)] = 1
		app.gameOver = false
		app.moveCount = 0
		app.moveHistory = make([]move.Move, 0)
		app.updateBoard()
		app.logMessage("Доска сброшена")

	case "switch":
		app.aiColor = oppositeColor(app.aiColor)
		app.logMessage(fmt.Sprintf("ИИ теперь играет за %v", app.aiColor))
		if app.aiColor == app.currentPlayerColor() {
			app.makeAIMove(4)
		}

	case "undo":
		if len(app.moveHistory) == 0 {
			app.logMessage("Нет ходов для отмены")
			return
		}
		lastMove := app.moveHistory[len(app.moveHistory)-1]
		app.moveHistory = app.moveHistory[:len(app.moveHistory)-1]
		app.currentBoard = board.NewBoard()
		app.positions = make(map[string]int)
		app.positions[boardToString(app.currentBoard)] = 1
		app.moveCount = 0
		for _, m := range app.moveHistory {
			move.MakeMove(&app.currentBoard, m)
			app.moveCount++
			app.positions[boardToString(app.currentBoard)]++
		}
		app.updateBoard()
		app.gameOver = false
		app.logMessage(fmt.Sprintf("Отменён ход: %s%d-%s%d", string('a'+lastMove.FromY), lastMove.FromX+1, string('a'+lastMove.ToY), lastMove.ToX+1))

	case "eval":
		score := evaluation.Evaluate(app.currentBoard)
		app.logMessage(fmt.Sprintf("Оценка позиции: %d (положительно для белых)", score))

	case "moves":
		currentColor := app.currentPlayerColor()
		moves := move.GenerateMoves(app.currentBoard, currentColor)
		if len(moves) == 0 {
			app.logMessage("Нет легальных ходов")
		} else {
			var moveList string
			for _, m := range moves {
				moveList += fmt.Sprintf("%s%d-%s%d ", string('a'+m.FromY), m.FromX+1, string('a'+m.ToY), m.ToX+1)
			}
			app.logMessage(fmt.Sprintf("Легальные ходы для %v: %s", currentColor, moveList))
		}

	case "exit":
		search.SaveData()
		app.window.Close()

	default:
		app.logMessage("Команды: move <depth>, play <move>, print, stats, reset, switch, undo, eval, moves, exit")
	}
}

func (app *ChessApp) makeAIMove(depth int) {
	if app.gameOver {
		app.infoLabel.SetText("Игра завершена. Используйте 'reset' для новой игры.")
		return
	}

	app.aiThinking = true
	app.infoLabel.SetText("ИИ думает...")
	app.logMessage(fmt.Sprintf("ИИ начинает поиск хода для %v на глубине %d", app.aiColor, depth))
	bestMove, stats := search.FindBestMove(app.currentBoard, depth, app.aiColor)
	app.lastStats = stats
	app.logMessage(fmt.Sprintf("ИИ нашёл ход: %s%d-%s%d, статистика: %d узлов, %v", string('a'+bestMove.FromY), bestMove.FromX+1, string('a'+bestMove.ToY), bestMove.ToX+1, stats.NodesEvaluated, stats.SearchTime))

	var message string
	if bestMove.FromX == 0 && bestMove.FromY == 0 && bestMove.ToX == 0 && bestMove.ToY == 0 {
		app.logMessage("ИИ не нашёл допустимых ходов")
		if move.IsKingInCheck(app.currentBoard, app.aiColor) {
			message = fmt.Sprintf("Мат! %v победили.", oppositeColor(app.aiColor))
		} else {
			message = "Пат! Ничья."
		}
	} else {
		if err := move.MakeMove(&app.currentBoard, bestMove); err != nil {
			app.logMessage(fmt.Sprintf("Ошибка хода ИИ: %v", err))
			message = "Ошибка ИИ: " + err.Error()
		} else {
			app.logMessage(fmt.Sprintf("Ход ИИ (%v): %s%d-%s%d", app.aiColor, string('a'+bestMove.FromY), bestMove.FromX+1, string('a'+bestMove.ToY), bestMove.ToX+1))
			app.playMoveSound()
			app.moveCount++
			app.moveHistory = append(app.moveHistory, bestMove)
			app.checkGameEnd(oppositeColor(app.aiColor))
			app.updateBoard()
			message = "ИИ сделал ход. Ваш ход."
		}
	}
	app.infoLabel.SetText(message)
	app.aiThinking = false
	if app.gameOver {
		search.SaveData()
	}
}

func (app *ChessApp) handleCellClick(x, y int) {
	if app.aiThinking {
		app.infoLabel.SetText("Подождите, ИИ думает...")
		return
	}
	if app.gameOver {
		app.infoLabel.SetText("Игра завершена. Используйте 'reset' для новой игры.")
		return
	}
	if app.aiColor == app.currentPlayerColor() {
		app.infoLabel.SetText("Ход ИИ. Используйте 'move <depth>' или 'switch'.")
		return
	}

	if app.selectedX == -1 {
		piece, color, err := app.currentBoard.GetPiece(x, y)
		if err != nil {
			app.logMessage(fmt.Sprintf("Ошибка при получении фигуры: %v", err))
			return
		}
		if piece != board.Empty && color != app.aiColor {
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
		if piece != board.Empty && color != app.aiColor {
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
			app.logMessage(fmt.Sprintf("Ход игрока: %s%d-%s%d", string('a'+app.selectedY), app.selectedX+1, string('a'+y), x+1))
			app.playMoveSound()
			app.selectedX, app.selectedY = -1, -1
			app.moveCount++
			app.moveHistory = append(app.moveHistory, m)
			app.checkGameEnd(app.aiColor)
			app.updateBoard()
			// app.logMessage(fmt.Sprintf("После хода игрока: moveCount=%d, currentColor=%v, aiColor=%v", app.moveCount, app.currentPlayerColor(), app.aiColor))
			// Автоматический ход ИИ
			if !app.gameOver && app.aiColor != oppositeColor(app.currentPlayerColor()) {
				app.logMessage("Вызываю makeAIMove для ИИ")
				app.makeAIMove(4) // Убрали go для синхронности
			}
		}
	}
}

// Остальные функции (logMessage, boardToString, etc.) остаются без изменений
func (app *ChessApp) logMessage(msg string) {
	log.Println(msg)
	app.logText.SetText(app.logText.Text + msg + "\n")
}

func boardToString(b board.Board) string {
	var s string
	for i := 7; i >= 0; i-- {
		s += fmt.Sprintf("%d | ", i+1)
		for j := 0; j < 8; j++ {
			piece, color, _ := b.GetPiece(i, j)
			if piece == board.Empty {
				s += ". "
			} else {
				symbol := pieceToSymbol(piece, color)
				s += symbol + " "
			}
		}
		s += "\n"
	}
	s += "  +-----------------\n"
	s += "    a b c d e f g h"
	return s
}

func pieceToSymbol(piece board.Piece, color board.Color) string {
	symbols := map[board.Piece]string{
		board.Pawn:   "P",
		board.Rook:   "R",
		board.Knight: "N",
		board.Bishop: "B",
		board.Queen:  "Q",
		board.King:   "K",
	}
	s := symbols[piece]
	if color == board.Black {
		s = strings.ToLower(s)
	}
	return s
}

func (app *ChessApp) playMoveSound() {
	go func() {
		file, err := os.Open("moveSound.mp3")
		if err != nil {
			app.logMessage("Файл moveSound.mp3 не найден")
			return
		}
		defer file.Close()

		streamer, _, err := mp3.Decode(file)
		if err != nil {
			app.logMessage(fmt.Sprintf("Ошибка декодирования MP3: %v", err))
			return
		}
		defer streamer.Close()

		speaker.Play(beep.Seq(streamer, beep.Callback(func() {})))
		time.Sleep(500 * time.Millisecond)
	}()
}

func (app *ChessApp) checkGameEnd(nextColor board.Color) {
	positionHash := boardToString(app.currentBoard)
	app.positions[positionHash]++
	if move.IsKingInCheck(app.currentBoard, nextColor) && app.isCheckmate(nextColor) {
		app.infoLabel.SetText(fmt.Sprintf("Мат! %v победили.", oppositeColor(nextColor)))
		app.logMessage(fmt.Sprintf("Игра завершена: мат %v", nextColor))
		app.gameOver = true
	} else if app.isCheckmate(nextColor) {
		app.infoLabel.SetText("Пат! Ничья.")
		app.logMessage("Игра завершена: пат")
		app.gameOver = true
	} else if app.positions[positionHash] >= 3 {
		app.infoLabel.SetText("Ничья по правилу трёхкратного повторения!")
		app.logMessage("Игра завершена: ничья по правилу трёхкратного повторения")
		app.gameOver = true
	}
}

func (app *ChessApp) currentPlayerColor() board.Color {
	if app.moveCount%2 == 0 {
		return board.White
	}
	return board.Black
}

func oppositeColor(color board.Color) board.Color {
	if color == board.White {
		return board.Black
	}
	return board.White
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

func (app *ChessApp) handleRightClick() {
	if app.gameOver {
		app.infoLabel.SetText("Игра завершена. Используйте 'reset' для новой игры.")
		return
	}
	if app.aiThinking {
		app.infoLabel.SetText("Подождите, ИИ думает...")
		return
	}
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
		return true // Пат
	}
	return false
}

func (app *ChessApp) updateBoard() {
	app.grid = app.createBoardGrid()
	app.window.SetContent(container.NewBorder(
		nil,
		container.NewVBox(
			app.infoLabel,
			container.NewMax(canvas.NewRectangle(color.RGBA{R: 30, G: 30, B: 30, A: 255}), app.logText),
			container.NewBorder(nil, nil, widget.NewLabel("Команда:"), nil, app.commandEntry),
		),
		nil,
		nil,
		app.grid,
	))
	app.window.Content().Refresh()
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
