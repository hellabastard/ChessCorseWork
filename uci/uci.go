package uci

import (
	"bufio"
	"chess-engine/board"
	"chess-engine/move"
	"chess-engine/search"
	"fmt"
	"os"
	"strings"
)

// StartUCI запускает UCI-протокол
func StartUCI() {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		switch {
		case text == "uci":
			handleUCI()
		case text == "isready":
			handleIsReady()
		case strings.HasPrefix(text, "position"):
			handlePosition(text)
		case strings.HasPrefix(text, "go"):
			handleGo()
		case text == "quit":
			return
		default:
			fmt.Println("Неизвестная команда:", text)
		}
	}
}

// handleUCI отвечает на команду "uci"
func handleUCI() {
	fmt.Println("id name MyChessEngine")
	fmt.Println("id author YourName")
	fmt.Println("uciok")
}

// handleIsReady отвечает на команду "isready"
func handleIsReady() {
	fmt.Println("readyok")
}

// handlePosition обрабатывает команду "position"
func handlePosition(command string) {
	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		return
	}

	// Обработка начальной позиции
	if parts[1] == "startpos" {
		currentBoard = board.NewBoard()
	} else if parts[1] == "fen" {
		// Пока не реализовано
		fmt.Println("FEN parsing not implemented")
		return
	}

	// Обработка ходов
	if len(parts) > 2 && parts[2] == "moves" {
		for i := 3; i < len(parts); i++ {
			m := parseMove(parts[i])
			if m == nil {
				fmt.Println("Ошибка при разборе хода:", parts[i])
				return
			}
			err := move.MakeMove(&currentBoard, *m)
			if err != nil {
				fmt.Println("Ошибка при выполнении хода:", err)
				return
			}
		}
	}
}

// handleGo обрабатывает команду "go" и запускает поиск лучшего хода
func handleGo() {
	result := search.FindBestMove(currentBoard, 3, board.Black) // Глубина поиска 3
	fmt.Printf("bestmove %s\n", formatMove(result))
}

// parseMove преобразует строку хода (например, "e2e4") в структуру Move
func parseMove(moveStr string) *move.Move {
	if len(moveStr) != 4 {
		return nil
	}

	fromX := int(moveStr[1] - '1')
	fromY := int(moveStr[0] - 'a')
	toX := int(moveStr[3] - '1')
	toY := int(moveStr[2] - 'a')

	return &move.Move{
		FromX: fromX,
		FromY: fromY,
		ToX:   toX,
		ToY:   toY,
	}
}

// formatMove преобразует структуру Move в строку хода (например, "e2e4")
func formatMove(m move.Move) string {
	fromFile := string('a' + rune(m.FromY))
	fromRank := string('1' + rune(m.FromX))
	toFile := string('a' + rune(m.ToY))
	toRank := string('1' + rune(m.ToX))
	return fmt.Sprintf("%s%s%s%s", fromFile, fromRank, toFile, toRank)
}

// currentBoard хранит текущее состояние доски
var currentBoard board.Board
