package search

import (
	"chess-engine/board"
	"chess-engine/evaluation"
	"chess-engine/move"
	"math"
)

// SearchResult содержит результат поиска (лучший ход и его оценку)
type SearchResult struct {
	BestMove move.Move
	Score    int
}

// Minimax реализует алгоритм минимакс с альфа-бета отсечением
func Minimax(b board.Board, depth int, alpha int, beta int, maximizingPlayer bool) SearchResult {
	// Базовый случай: если достигнута максимальная глубина или игра окончена
	if depth == 0 {
		return SearchResult{
			Score: evaluation.Evaluate(b),
		}
	}

	var bestMove move.Move
	var bestScore int

	if maximizingPlayer {
		bestScore = math.MinInt // Инициализируем худшим возможным значением для максимизирующего игрока
	} else {
		bestScore = math.MaxInt // Инициализируем худшим возможным значением для минимизирующего игрока
	}

	// Генерация всех возможных ходов
	moves := move.GenerateMoves(b, board.White) // Или board.Black, в зависимости от maximizingPlayer
	for _, m := range moves {
		// Создаем копию доски и выполняем ход
		newBoard := b
		err := move.MakeMove(&newBoard, m)
		if err != nil {
			continue // Пропускаем некорректные ходы
		}

		// Рекурсивный вызов для оценки хода
		result := Minimax(newBoard, depth-1, alpha, beta, !maximizingPlayer)

		// Обновляем лучший ход и оценку
		if maximizingPlayer {
			if result.Score > bestScore {
				bestScore = result.Score
				bestMove = m
			}
			alpha = max(alpha, bestScore)
		} else {
			if result.Score < bestScore {
				bestScore = result.Score
				bestMove = m
			}
			beta = min(beta, bestScore)
		}

		// Альфа-бета отсечение
		if beta <= alpha {
			break
		}
	}

	return SearchResult{
		BestMove: bestMove,
		Score:    bestScore,
	}
}

// FindBestMove находит лучший ход для текущей позиции
func FindBestMove(b board.Board, depth int, boardColor board.Color) move.Move {
	// Если boardColor == Black, ИИ ходит за черных (минимизация)
	// Если boardColor == White, ИИ ходит за белых (максимизация)
	maximizingPlayer := (boardColor == board.White)
	result := Minimax(b, depth, math.MinInt, math.MaxInt, maximizingPlayer)
	return result.BestMove
}

// Вспомогательные функции для нахождения максимума и минимума
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
