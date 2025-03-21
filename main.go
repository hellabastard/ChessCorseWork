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
var flagArray []int = []int{0, 1}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Создаем папку logs, если она не существует
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		log.Fatalf("Ошибка создания папки logs: %v", err)
	}

	// Определяем следующий номер файла
	files, err := filepath.Glob("logs/log*.txt")
	if err != nil {
		log.Fatalf("Ошибка чтения файлов логов: %v", err)
	}
	gameCounter = len(files) + 1 // Номер следующего файла

	// Инициализация звука
	speaker.Init(44100, 44100/10) // Частота 44100 Гц, буфер на 100 мс
}

func handleConsoleCommands(app *ui.ChessApp) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Split(input, " ")

		switch parts[0] {
		case "pause":
			app.Pause()

		case "depth=":
			if len(parts) < 2 {
				log.Println("Ошибка: укажите глубину")
			} else {
				depth, err := strconv.Atoi(parts[1])
				if err != nil || depth <= 0 {
					log.Println("Ошибка: глубина должна быть положительная")
				} else {
					app.SetAIDepth(depth)
				}
			}

		case "reset":
			app.Reset()

		case "eval":
			app.PrintLastMoveEval()

		case "help":
			log.Println("pause, help, depth= <value>, reset, eval, exit= <flag>")

		case "exit=":
			if len(parts) < 2 {
				log.Println("Укажите флаг выхода")
			} else {
				flag, err := strconv.Atoi(parts[1])
				if err != nil || !(contains(flagArray, 0)) || !(contains(flagArray, 1)) {
					log.Println("Ошибка! Не найден флаг")
				} else {
					app.Exit(flag)
				}
			}

		case "print":
			//выводим доску в консоль

		default:
			log.Println("Неверная команда")
		}
	}
}

func main() {
	// Открываем файл логов
	logFile, err := os.Create(filepath.Join("logs", "log"+strconv.Itoa(gameCounter)+".txt"))
	if err != nil {
		log.Fatalf("Ошибка создания файла логов: %v", err)
	}
	defer logFile.Close()

	// Настраиваем вывод в консоль (os.Stdout) и файл
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// Выводим заставку
	fmt.Println("Курсовая работа на тему: игра Шахматы\nВыполнил: студент группы 24ВВВ1 Будников А.С.\nПриняла: к.т.н. доцент Генералова А.А.")
	// Ожидаем нажатия Enter для начала игры
	fmt.Printf("\nДля запуска игры нажмите Enter...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	// Запускаем приложение
	chessApp := ui.NewChessApp()
	go handleConsoleCommands(chessApp)
	chessApp.Run()
}

func contains(arr []int, value int) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}
