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
		board.Black: color.Black,
		board.White: color.White,
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
	positions            map[string]int // История позиций для правила трёхкратного повторения
	gameOver             bool           // Флаг окончания игры
	aiThinking           bool           // Флаг, показывающий, что ИИ думает
	moveCount            int            // Счётчик ходов для определения первого хода
	paused               bool
	aiDepth              int
}

func NewChessApp() *ChessApp {
	// Загружаем данные ИИ при создании приложения
	search.LoadData()

	app := &ChessApp{
		currentBoard: board.NewBoard(),
		selectedX:    -1,
		selectedY:    -1,
		logText:      widget.NewEntry(),
		positions:    make(map[string]int),
		gameOver:     false,
		aiThinking:   false,
		moveCount:    0,
		paused:       false,
		aiDepth:      5,
	}
	app.positions[boardToString(app.currentBoard)] = 1
	return app
}

func (appl *ChessApp) Run() {
	myApp := app.New()
	appl.window = myApp.NewWindow("Шахматы")

	appl.grid = appl.createBoardGrid()
	appl.infoLabel = widget.NewLabel("Ваш ход. Выберите фигуру.")

	// Настраиваем logText
	appl.logText.MultiLine = true
	appl.logText.Wrapping = fyne.TextWrapWord
	appl.logText.Disable() // Используем Disable вместо SetReadOnly для Fyne 2.5.4

	// Оборачиваем logText в контейнер с тёмным фоном
	logContainer := container.NewMax(
		canvas.NewRectangle(color.RGBA{R: 30, G: 30, B: 30, A: 255}),
		appl.logText,
	)

	content := container.NewBorder(
		nil,
		container.NewVBox(appl.infoLabel, logContainer),
		nil,
		nil,
		appl.grid,
	)
	appl.window.SetContent(content)
	appl.window.Resize(fyne.NewSize(cellSize*8, cellSize*8+100))

	// Сохраняем данные ИИ при закрытии окна
	appl.window.SetCloseIntercept(func() {
		search.SaveData()
		appl.window.Close()
	})

	appl.window.Show()
	myApp.Run()
}

func (app *ChessApp) logMessage(msg string) {
	log.Println(msg)
	app.logText.SetText(app.logText.Text + msg + "\n")
}

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

func (app *ChessApp) playMoveSound() {
	go func() {
		file, err := os.Open("moveSound.mp3")
		if err != nil {
			app.logMessage("Файл moveSound.mp3 не найден, воспроизводим тон")
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

func (app *ChessApp) handleCellClick(x, y int) {
	if app.aiThinking && app.moveCount > 0 {
		app.infoLabel.SetText("Подождите, ИИ думает...")
		return
	}
	if app.gameOver {
		app.infoLabel.SetText("Игра завершена. Начните новую игру.")
		return
	}

	if app.selectedX == -1 {
		piece, color, err := app.currentBoard.GetPiece(x, y)
		if err != nil {
			app.logMessage(fmt.Sprintf("Ошибка при получении фигуры: %v", err))
			return
		}
		if piece != board.Empty && color == board.White && app.paused == false {
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
			app.playMoveSound()
			app.selectedX, app.selectedY = -1, -1
			app.moveCount++
			positionHash := boardToString(app.currentBoard)
			app.positions[positionHash]++
			app.updateBoard()

			if move.IsKingInCheck(app.currentBoard, board.Black) && app.isCheckmate(board.Black) {
				app.infoLabel.SetText("Мат! Белые победили.")
				app.logMessage("Игра завершена: мат чёрным. Победитель: Белые")
				app.gameOver = true
				search.SaveData()
				return
			} else if app.isCheckmate(board.Black) {
				app.infoLabel.SetText("Пат! Ничья.")
				app.logMessage("Игра завершена: пат для чёрных")
				app.gameOver = true
				search.SaveData()
				return
			} else if app.positions[positionHash] >= 3 {
				app.infoLabel.SetText("Ничья по правилу трёхкратного повторения!")
				app.logMessage("Игра завершена: ничья по правилу трёхкратного повторения")
				app.gameOver = true
				search.SaveData()
				return
			}

			app.makeAIMove()
		}
	}
}

func (app *ChessApp) makeAIMove() {
	if app.gameOver {
		app.infoLabel.SetText("Игра завершена. Начните новую игру.")
		return
	}

	app.aiThinking = true
	app.infoLabel.SetText("ИИ думает...")
	go func() {
		bestMove, _ := search.FindBestMove(app.currentBoard, app.aiDepth, board.Black)
		var message string
		if bestMove.FromX == 0 && bestMove.FromY == 0 && bestMove.ToX == 0 && bestMove.ToY == 0 {
			app.logMessage("ИИ не нашёл допустимых ходов")
			if move.IsKingInCheck(app.currentBoard, board.Black) {
				message = "Мат! Белые победили."
			} else {
				message = "Пат! Ничья."
			}
		} else {
			if err := move.MakeMove(&app.currentBoard, bestMove); err != nil {
				app.logMessage(fmt.Sprintf("Ошибка при выполнении хода ИИ: %v", err))
				message = "Ошибка ИИ: " + err.Error()
			} else {
				app.logMessage(fmt.Sprintf("Ход ИИ (чёрные): %s%d-%s%d", string('a'+bestMove.FromY), bestMove.FromX+1, string('a'+bestMove.ToY), bestMove.ToX+1))
				app.playMoveSound()
				app.moveCount++
				positionHash := boardToString(app.currentBoard)
				app.positions[positionHash]++
				app.updateBoard()

				if move.IsKingInCheck(app.currentBoard, board.White) && app.isCheckmate(board.White) {
					message = "Мат! Чёрные победили."
				} else if app.isCheckmate(board.White) {
					message = "Пат! Ничья."
				} else if app.positions[positionHash] >= 3 {
					message = "Ничья по правилу трёхкратного повторения!"
				} else {
					message = "ИИ сделал ход. Ваш ход."
				}
			}
		}
		app.infoLabel.SetText(message)
		app.aiThinking = false
		app.gameOver = message != "ИИ сделал ход. Ваш ход."
		if app.gameOver {
			search.SaveData()
		}
	}()
}

func (app *ChessApp) createCell(x, y int) fyne.CanvasObject {
	lightColor := color.RGBA{R: 240, G: 217, B: 181, A: 255}
	darkColor := color.RGBA{R: 181, G: 136, B: 99, A: 255}

	cellColor := darkColor
	if (x+y)%2 == 1 {
		cellColor = lightColor
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
		app.infoLabel.SetText("Игра завершена. Начните новую игру.")
		return
	}
	if app.aiThinking && app.moveCount > 0 {
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
			return true
		}
		return true
	}
	return false
}

func (app *ChessApp) updateBoard() {
	app.grid = app.createBoardGrid()
	app.window.SetContent(container.NewBorder(nil, container.NewVBox(app.infoLabel, container.NewMax(canvas.NewRectangle(color.RGBA{R: 30, G: 30, B: 30, A: 255}), app.logText)), nil, nil, app.grid))
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

// Консольные команды
func (app *ChessApp) PrintLastMoveEval() {
	if app.moveCount == 0 {
		log.Println("Нет ходов для оценки")
		return
	}
	score := evaluation.Evaluate(app.currentBoard)
	log.Printf("Оценка позиции: %d (положительно для белых)", score)
}

func (app *ChessApp) Pause() {
	if app.aiThinking {
		log.Println("Невозможно поставить на паузу. ИИ делает ход")
		return
	}
	app.paused = !app.paused
	if app.paused {
		log.Println("Игра приостановлена")
		app.infoLabel.SetText("Игра приостановлена")
	} else {
		log.Println("Игра возобновлена")
		app.infoLabel.SetText("Ваш ход. Выберите фигуру.")
	}
}

func (app *ChessApp) SetAIDepth(depth int) {
	app.aiDepth = depth
	log.Printf("Глубина поиска ИИ установлена на %d", depth)
}

func (app *ChessApp) Reset() {
	app.currentBoard = board.NewBoard()
	app.selectedX, app.selectedY = -1, -1
	app.positions = make(map[string]int)
	app.positions[boardToString(app.currentBoard)] = 1
	app.gameOver = false
	app.aiThinking = false
	app.moveCount = 0
	app.paused = false
	app.updateBoard()
	app.infoLabel.SetText("Игра сброшена. Ваш ход.")
	log.Println("Игра сброшена")
}

func (app *ChessApp) Exit(flag int) {
	if flag == 0 {
		app.window.Close()
		return
	} else {
		search.SaveData()
		app.window.Close()
	}
}
