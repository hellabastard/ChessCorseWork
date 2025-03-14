package main

import (
	"bufio"
	"chess-engine/ui"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var gameCounter int

func init() {
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
	gameCounter = len(files) // Номер следующего файла
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
	fmt.Println("Курсовая работа на тему: игра Шахматы\nВыполнил: студент группы 24ВВВ1 Будников А.С.\nПриняла: к.т.н. доцент Генералова А.А.\n")
	// Ожидаем нажатия Enter для начала игры
	fmt.Printf("Для запуска игры нажмите Enter...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	// Запускаем приложение
	chessApp := ui.NewChessApp()
	chessApp.Run()
}
