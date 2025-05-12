package main

import (
	"bufio"
	"chess-engine/search"
	"chess-engine/ui"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var gameCounter int
var flagArray []int = []int{0, 1}

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
				flagStr := parts[1]
				flag, err := strconv.Atoi(flagStr)
				if err != nil || !(contains(flagArray, flag)) {
					log.Println("Ошибка! Не найден флаг")
				} else {
					app.Exit(flag)
				}
			}
		default:
			log.Println("Неверная команда")
		}
	}
}

func main() {
	logFile, err := os.Create(filepath.Join("logs", "log"+strconv.Itoa(gameCounter)+".txt"))
	if err != nil {
		log.Fatalf("Ошибка создания файла логов: %v", err)
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	reader := bufio.NewReader(os.Stdin)
	chessApp := ui.NewChessApp()

	fmt.Println("\nКурсовая работа на тему: игра Шахматы\nВыполнил: студент группы 24ВВВ1 Будников А.С.\nПриняла: к.т.н. доцент Генералова А.А.\n\nНажмите Enter для запуска меню")
	_, _ = reader.ReadString('\n')

mainLoop:
	for {
		fmt.Println("Выберите один из пунктов меню\n1. Начать игру\n2. Настройки\n3. Выход")
		fmt.Print("\nМой выбор: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка при чтении ввода:", err)
			continue
		}

		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			go handleConsoleCommands(chessApp)
			chessApp.Run()
			return
		case "2":
			for {
				fmt.Println("\n=== МЕНЮ НАСТРОЕК ===")
				fmt.Printf("1. Глубина поиска AI (текущая: %d)\n", chessApp.GetAiDepth())
				fmt.Printf("2. Максимальное время поиска AI (текущая: %d)\n", search.GetTimeLimit())
				fmt.Println("3. Вернуться в главное меню")
				fmt.Print("\nМой выбор: ")

				input, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Ошибка при чтении ввода:", err)
					continue
				}

				choice := strings.TrimSpace(input)
				switch choice {
				case "1":
					fmt.Print("Введите новую глубину поиска AI: ")
					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Println("Ошибка при чтении ввода:", err)
						continue
					}
					depth, err := strconv.Atoi(strings.TrimSpace(input))
					if err != nil || depth <= 0 {
						fmt.Println("Ошибка: глубина должна быть положительным числом")
						continue
					} else {
						chessApp.SetAIDepth(depth)
						fmt.Printf("Глубина поиска AI установлена: %d\n", depth)
						continue
					}
				case "2":
					fmt.Print("Введите новое максимальное время поиска AI: ")
					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Println("Ошибка при чтении ввода:", err)
						continue
					}
					timer, err := strconv.Atoi(strings.TrimSpace(input))
					if err != nil || timer <= 0 {
						fmt.Println("Ошибка: время должно быть положительным числом")
						continue
					} else {
						search.SetTimeLimit(timer)
						fmt.Printf("Максимальное время поиска AI установлена: %d\n", search.GetTimeLimit())
						continue
					}
				case "3":
					continue mainLoop
				default:
					fmt.Println("Неверный выбор, попробуйте снова.")
				}
			}
		case "3":
			fmt.Println("Выход из программы")
			return
		default:
			fmt.Println("Неверный выбор, попробуйте снова.")
			continue
		}
	}
}

func contains(arr []int, value int) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}
