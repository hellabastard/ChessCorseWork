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
	_, _ = reader.ReadString('\n')

	chessApp := ui.NewChessApp()
	chessApp.Run()
}
