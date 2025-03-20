package main

import (
	"bufio"
	"chess-engine/ui"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/faiface/beep/speaker"
)

var gameCounter int

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		log.Fatalf("Ошибка создания папки logs: %v", err)
	}

	files, err := filepath.Glob("logs/log*.txt")
	if err != nil {
		log.Fatalf("Ошибка чтения файлов логов: %v", err)
	}
	gameCounter = len(files) + 1

	speaker.Init(44100, 44100/10)
}

func main() {
	logFile, err := os.Create(filepath.Join("logs", "log"+strconv.Itoa(gameCounter)+".txt"))
	if err != nil {
		log.Fatalf("Ошибка создания файла логов: %v", err)
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	fmt.Println("Курсовая работа на тему: игра Шахматы\nВыполнил: студент группы 24ВВВ1 Будников А.С.\nПриняла: к.т.н. доцент Генералова А.А.")
	fmt.Printf("\nДля запуска нажмите Enter...")
	reader := bufio.NewReader(os.Stdin)
	_, err = reader.ReadString('\n')
	if err != nil {
		log.Printf("Ошибка чтения ввода: %v", err)
		return
	}

	chessApp := ui.NewChessApp()
	go handleConsoleCommands(chessApp) // Запускаем обработку консоли в отдельной горутине
	chessApp.Run()
}

func handleConsoleCommands(app *ui.ChessApp) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Введите команду (stats, eval, switch, pause, depth, reset): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Split(input, " ")

		switch parts[0] {
		case "stats":
			app.PrintStats()

		case "eval":
			app.PrintLastMoveEval()

		case "switch":
			app.SwitchAIColor()

		case "pause":
			app.TogglePause()

		case "depth":
			if len(parts) < 2 {
				log.Println("Ошибка: укажите глубину, например 'depth 4'")
			} else {
				depth, err := strconv.Atoi(parts[1])
				if err != nil || depth <= 0 {
					log.Println("Ошибка: глубина должна быть положительным числом")
				} else {
					app.SetAIDepth(depth)
				}
			}

		case "reset":
			app.ResetGame()

		default:
			log.Println("Команды: stats, eval, switch, pause, depth <value>, reset")
		}
	}
}
