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
	positions            map[string]int
	gameOver             bool
	aiThinking           bool
	moveCount            int
	aiColor              board.Color
	lastStats            search.SearchStats
	moveHistory          []move.Move
	paused               bool
	aiDepth              int
}

func NewChessApp() *ChessApp {
	search.LoadData()
	app := &ChessApp{
		currentBoard: board.NewBoard(),
		selectedX:    -1,
		selectedY:    -1,
		positions:    make(map[string]int),
		gameOver:     false,
		aiThinking:   false,
		moveCount:    0,
		aiColor:      board.Black,
		moveHistory:  make([]move.Move, 0),
		paused:       false,
		aiDepth:      4, // Глубина по умолчанию
	}
	app.positions[boardToString(app.currentBoard)] = 1
	return app
}

func (appl *ChessApp) Run() {
	myApp := app.New()
	appl.window = myApp.NewWindow("Шахматы")

	appl.grid = appl.createBoardGrid()
	appl.infoLabel = widget.NewLabel("Ваш ход. Выберите фигуру.")
	content := container.NewBorder(
		nil,
		appl.infoLabel,
		nil,
		nil,
		appl.grid,
	)
	appl.window.SetContent(content)
	appl.window.Resize(fyne.NewSize(cellSize*8, cellSize*8+50))

	appl.window.SetCloseIntercept(func() {
		search.SaveData()
		appl.window.Close()
	})

	appl.window.Show()
	myApp.Run()
}

func (app *ChessApp) handleCellClick(x, y int) {
	if app.aiThinking {
		app.infoLabel.SetText("Подождите, ИИ думает...")
		return
	}
	if app.gameOver {
		app.infoLabel.SetText("Игра завершена.")
		return
	}
	if app.paused {
		app.infoLabel.SetText("Игра на паузе. Используйте 'pause' в консоли для продолжения.")
		return
	}
	if app.aiColor == app.currentPlayerColor() {
		app.infoLabel.SetText("Пододждите. ИИ думает...")
		return
	}

	if app.selectedX == -1 {
		piece, color, err := app.currentBoard.GetPiece(x, y)
		if err != nil {
			log.Printf("Ошибка при получении фигуры: %v", err)
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
			log.Printf("Ошибка при получении фигуры: %v", err)
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
		capturedPiece, capturedColor, _ := app.currentBoard.GetPiece(x, y)
		m.Captured = capturedPiece
		m.CapturedColor = capturedColor
		if err := move.MakeMove(&app.currentBoard, m); err != nil {
			app.infoLabel.SetText("Некорректный ход: " + err.Error())
		} else {
			log.Printf("Ход игрока: %s%d-%s%d", string('a'+app.selectedY), app.selectedX+1, string('a'+y), x+1)
			app.playMoveSound()
			app.selectedX, app.selectedY = -1, -1
			app.moveCount++
			app.moveHistory = append(app.moveHistory, m)
			app.checkGameEnd(app.aiColor)
			app.updateBoard()
			if !app.gameOver && !app.paused && app.aiColor != oppositeColor(app.currentPlayerColor()) {
				log.Printf("После хода игрока: moveCount=%d, currentColor=%v, aiColor=%v, вызываю ход ИИ", app.moveCount, app.currentPlayerColor(), app.aiColor)
				app.makeAIMove(app.aiDepth) // Убрали go для синхронности
			}
		}
	}
}

func (app *ChessApp) makeAIMove(depth int) {
	if app.gameOver || app.paused {
		return
	}

	app.aiThinking = true
	app.infoLabel.SetText("ИИ думает...")
	log.Printf("ИИ думает... (глубина=%d)", depth)

	bestMove, stats := search.FindBestMove(app.currentBoard, depth, app.aiColor)
	app.lastStats = stats

	if bestMove.FromX == 0 && bestMove.FromY == 0 && bestMove.ToX == 0 && bestMove.ToY == 0 {
		log.Printf("ИИ не нашёл допустимых ходов")
		if move.IsKingInCheck(app.currentBoard, app.aiColor) {
			app.infoLabel.SetText(fmt.Sprintf("Мат! %v победили.", oppositeColor(app.aiColor)))
		} else {
			app.infoLabel.SetText("Пат! Ничья.")
		}
		app.gameOver = true
	} else {
		capturedPiece, capturedColor, _ := app.currentBoard.GetPiece(bestMove.ToX, bestMove.ToY)
		bestMove.Captured = capturedPiece
		bestMove.CapturedColor = capturedColor
		if bestMove.FromX == bestMove.ToX && (bestMove.FromY-bestMove.ToY == 2 || bestMove.ToY-bestMove.FromY == 2) {
			bestMove.IsCastling = true
		}

		if err := move.MakeMove(&app.currentBoard, bestMove); err != nil {
			log.Printf("Ошибка хода ИИ: %v", err)
			app.infoLabel.SetText("Ошибка ИИ: " + err.Error())
		} else {
			log.Printf("Ход ИИ (%v): %s%d-%s%d", app.aiColor, string('a'+bestMove.FromY), bestMove.FromX+1, string('a'+bestMove.ToY), bestMove.ToX+1)
			app.playMoveSound()
			app.moveCount++
			app.moveHistory = append(app.moveHistory, bestMove)
			app.checkGameEnd(oppositeColor(app.aiColor))
			app.updateBoard()
			app.infoLabel.SetText("ИИ сделал ход. Ваш ход.")
		}
	}
	app.aiThinking = false
	if app.gameOver {
		search.SaveData()
	}
}

// Команды консоли
func (app *ChessApp) PrintStats() {
	if app.lastStats.NodesEvaluated == 0 {
		log.Println("Статистика ИИ недоступна: ИИ ещё не делал ходов")
	} else {
		log.Printf("Статистика ИИ: Оценено узлов: %d, Время поиска: %v, Скорость: %.2f узлов/с",
			app.lastStats.NodesEvaluated, app.lastStats.SearchTime, float64(app.lastStats.NodesEvaluated)/app.lastStats.SearchTime.Seconds())
	}
	log.Printf("Статистика игрока: Сделано ходов: %d", app.moveCount/2)
}

func (app *ChessApp) PrintLastMoveEval() {
	if len(app.moveHistory) == 0 {
		log.Println("Нет ходов для оценки")
		return
	}
	lastMove := app.moveHistory[len(app.moveHistory)-1]
	tempBoard := app.currentBoard
	move.UndoMove(&tempBoard, lastMove)
	score := evaluation.Evaluate(tempBoard)
	log.Printf("Оценка позиции перед последним ходом %s%d-%s%d: %d (положительно для белых)", string('a'+lastMove.FromY), lastMove.FromX+1, string('a'+lastMove.ToY), lastMove.ToX+1, score)
}

func (app *ChessApp) SwitchAIColor() {
	app.aiColor = oppositeColor(app.aiColor)
	log.Printf("ИИ теперь играет за %v", app.aiColor)
	if !app.gameOver && !app.paused && app.aiColor == app.currentPlayerColor() {
		app.makeAIMove(app.aiDepth)
	}
}

func (app *ChessApp) TogglePause() {
	app.paused = !app.paused
	if app.paused {
		log.Println("Игра приостановлена")
		app.infoLabel.SetText("Игра на паузе")
	} else {
		log.Println("Игра возобновлена")
		app.infoLabel.SetText("Ваш ход. Выберите фигуру.")
		if !app.gameOver && app.aiColor == app.currentPlayerColor() {
			app.makeAIMove(app.aiDepth)
		}
	}
}

func (app *ChessApp) SetAIDepth(depth int) {
	app.aiDepth = depth
	log.Printf("Глубина поиска ИИ установлена на %d", depth)
}

func (app *ChessApp) ResetGame() {
	app.currentBoard = board.NewBoard()
	app.selectedX, app.selectedY = -1, -1
	app.positions = make(map[string]int)
	app.positions[boardToString(app.currentBoard)] = 1
	app.gameOver = false
	app.aiThinking = false
	app.moveCount = 0
	app.moveHistory = make([]move.Move, 0)
	app.paused = false
	app.updateBoard()
	app.infoLabel.SetText("Игра сброшена. Ваш ход.")
	log.Println("Игра сброшена")
	if app.aiColor == app.currentPlayerColor() {
		app.makeAIMove(app.aiDepth)
	}
}

func (app *ChessApp) playMoveSound() {
	go func() {
		file, err := os.Open("moveSound.mp3")
		if err != nil {
			log.Printf("Файл moveSound.mp3 не найден")
			return
		}
		defer file.Close()

		streamer, _, err := mp3.Decode(file)
		if err != nil {
			log.Printf("Ошибка декодирования MP3: %v", err)
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
		log.Printf("Игра завершена: мат %v", nextColor)
		app.gameOver = true
	} else if app.isCheckmate(nextColor) {
		app.infoLabel.SetText("Пат! Ничья.")
		log.Printf("Игра завершена: пат")
		app.gameOver = true
	} else if app.positions[positionHash] >= 3 {
		app.infoLabel.SetText("Ничья по правилу трёхкратного повторения!")
		log.Printf("Игра завершена: ничья по правилу трёхкратного повторения")
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
		app.infoLabel.SetText("Игра завершена.")
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
		app.infoLabel,
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
