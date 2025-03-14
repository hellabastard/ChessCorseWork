// package search

// import (
// 	"chess-engine/board"
// 	"chess-engine/evaluation"
// 	"chess-engine/move"
// 	"fmt"
// 	"math"
// 	"sort"
// )

// // TranspositionTable для кэширования результатов
// var transpositionTable = make(map[string]SearchResult)

// type SearchResult struct {
// 	BestMove move.Move
// 	Score    int
// }

// func Minimax(b board.Board, depth int, alpha int, beta int, maximizingPlayer bool) SearchResult {
// 	// Хэш позиции для таблицы транспозиций
// 	hash := boardToString(b)
// 	if result, ok := transpositionTable[hash]; ok && depth <= result.Score {
// 		return result
// 	}

// 	// Базовый случай: достигнута максимальная глубина
// 	if depth == 0 {
// 		score := evaluation.Evaluate(b)
// 		return SearchResult{Score: score}
// 	}

// 	var bestMove move.Move
// 	var bestScore int

// 	if maximizingPlayer {
// 		bestScore = math.MinInt // Худшее значение для максимизации
// 	} else {
// 		bestScore = math.MaxInt // Худшее значение для минимизации
// 	}

// 	// Генерация ходов
// 	var moves []move.Move
// 	if maximizingPlayer {
// 		moves = move.GenerateMoves(b, board.White)
// 	} else {
// 		moves = move.GenerateMoves(b, board.Black)
// 	}

// 	// Если ходов нет, возвращаем текущую оценку
// 	if len(moves) == 0 {
// 		return SearchResult{
// 			BestMove: move.Move{},
// 			Score:    evaluation.Evaluate(b),
// 		}
// 	}

// 	// Сортировка ходов для улучшения альфа-бета отсечения
// 	sortMoves(moves, b, maximizingPlayer)

// 	for _, m := range moves {
// 		newBoard := b
// 		if err := move.MakeMove(&newBoard, m); err != nil {
// 			continue // Пропускаем некорректные ходы
// 		}

// 		// Рекурсивный вызов
// 		result := Minimax(newBoard, depth-1, alpha, beta, !maximizingPlayer)

// 		if maximizingPlayer {
// 			if result.Score > bestScore {
// 				bestScore = result.Score
// 				bestMove = m
// 			}
// 			alpha = max(alpha, bestScore)
// 		} else {
// 			if result.Score < bestScore {
// 				bestScore = result.Score
// 				bestMove = m
// 			}
// 			beta = min(beta, bestScore)
// 		}

// 		// Альфа-бета отсечение
// 		if beta <= alpha {
// 			break
// 		}
// 	}

// 	// Сохраняем результат в таблицу транспозиций
// 	result := SearchResult{
// 		BestMove: bestMove,
// 		Score:    bestScore,
// 	}
// 	transpositionTable[hash] = result
// 	return result
// }

// func FindBestMove(b board.Board, depth int, boardColor board.Color) move.Move {
// 	maximizingPlayer := (boardColor == board.White)
// 	result := Minimax(b, depth, math.MinInt, math.MaxInt, maximizingPlayer)
// 	return result.BestMove
// }

// func max(a, b int) int {
// 	if a > b {
// 		return a
// 	}
// 	return b
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

// // sortMoves сортирует ходы с приоритетом: шахи, взятия, продвижение пешек
// func sortMoves(moves []move.Move, b board.Board, maximizingPlayer bool) {
// 	sort.Slice(moves, func(i, j int) bool {
// 		moveI, moveJ := moves[i], moves[j]
// 		pieceI, _, _ := b.GetPiece(moveI.FromX, moveI.FromY)
// 		pieceJ, _, _ := b.GetPiece(moveJ.FromX, moveJ.FromY)
// 		targetPieceI, _, _ := b.GetPiece(moveI.ToX, moveI.ToY)
// 		targetPieceJ, _, _ := b.GetPiece(moveJ.ToX, moveJ.ToY)

// 		// Проверяем, дает ли ход шах
// 		newBoardI := b
// 		newBoardJ := b
// 		isCheckI := move.MakeMove(&newBoardI, moveI) == nil && move.IsKingInCheck(newBoardI, board.White)
// 		isCheckJ := move.MakeMove(&newBoardJ, moveJ) == nil && move.IsKingInCheck(newBoardJ, board.White)
// 		if !maximizingPlayer {
// 			isCheckI = move.IsKingInCheck(newBoardI, board.Black)
// 			isCheckJ = move.IsKingInCheck(newBoardJ, board.Black)
// 		}

// 		// Приоритеты: шах = 100, взятие = 10, продвижение пешки = 5
// 		scoreI := 0
// 		if isCheckI {
// 			scoreI += 100
// 		}
// 		if targetPieceI != board.Empty {
// 			scoreI += 10
// 		}
// 		if pieceI == board.Pawn && (moveI.ToX == 0 || moveI.ToX == 7) {
// 			scoreI += 5
// 		}

// 		scoreJ := 0
// 		if isCheckJ {
// 			scoreJ += 100
// 		}
// 		if targetPieceJ != board.Empty {
// 			scoreJ += 10
// 		}
// 		if pieceJ == board.Pawn && (moveJ.ToX == 0 || moveJ.ToX == 7) {
// 			scoreJ += 5
// 		}

// 		return scoreI > scoreJ
// 	})
// }

// // boardToString создает строковый хэш доски для таблицы транспозиций
// func boardToString(b board.Board) string {
// 	var s string
// 	for i := 0; i < 8; i++ {
// 		for j := 0; j < 8; j++ {
// 			piece, color, _ := b.GetPiece(i, j)
// 			s += fmt.Sprintf("%d%d", piece, color)
// 		}
// 	}
// 	return s
// }

package search

import (
	"chess-engine/board"
	"chess-engine/evaluation"
	"chess-engine/move"
	"fmt"
	"math"
	"math/rand" // Добавляем для случайности
	"sort"
	"time" // Для инициализации рандома
)

var transpositionTable = make(map[string]SearchResult)

type SearchResult struct {
	BestMove move.Move
	Score    int
}

func init() {
	rand.Seed(time.Now().UnixNano()) // Инициализируем генератор случайных чисел
}

func Minimax(b board.Board, depth int, alpha int, beta int, maximizingPlayer bool) SearchResult {
	hash := boardToString(b)
	if result, ok := transpositionTable[hash]; ok && depth <= result.Score {
		return result
	}

	if depth == 0 {
		score := evaluation.Evaluate(b)
		return SearchResult{Score: score}
	}

	var bestMove move.Move
	var bestScore int

	if maximizingPlayer {
		bestScore = math.MinInt
	} else {
		bestScore = math.MaxInt
	}

	var moves []move.Move
	if maximizingPlayer {
		moves = move.GenerateMoves(b, board.White)
	} else {
		moves = move.GenerateMoves(b, board.Black)
	}

	if len(moves) == 0 {
		return SearchResult{
			BestMove: move.Move{},
			Score:    evaluation.Evaluate(b),
		}
	}

	sortMoves(moves, b, maximizingPlayer)

	for _, m := range moves {
		newBoard := b
		if err := move.MakeMove(&newBoard, m); err != nil {
			continue
		}

		result := Minimax(newBoard, depth-1, alpha, beta, !maximizingPlayer)

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
			// beta = max(beta, bestScore)
		}

		if beta <= alpha {
			break
		}
	}

	result := SearchResult{
		BestMove: bestMove,
		Score:    bestScore,
	}
	transpositionTable[hash] = result
	return result
}

func FindBestMove(b board.Board, depth int, boardColor board.Color) move.Move {
	maximizingPlayer := (boardColor == board.White)
	result := Minimax(b, depth, math.MinInt, math.MaxInt, maximizingPlayer)
	return result.BestMove
}

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

func sortMoves(moves []move.Move, b board.Board, maximizingPlayer bool) {
	sort.Slice(moves, func(i, j int) bool {
		moveI, moveJ := moves[i], moves[j]
		pieceI, _, _ := b.GetPiece(moveI.FromX, moveI.FromY)
		pieceJ, _, _ := b.GetPiece(moveJ.FromX, moveJ.FromY)
		targetPieceI, _, _ := b.GetPiece(moveI.ToX, moveI.ToY)
		targetPieceJ, _, _ := b.GetPiece(moveJ.ToX, moveJ.ToY)

		newBoardI := b
		newBoardJ := b
		isCheckI := move.MakeMove(&newBoardI, moveI) == nil && move.IsKingInCheck(newBoardI, board.White)
		isCheckJ := move.MakeMove(&newBoardJ, moveJ) == nil && move.IsKingInCheck(newBoardJ, board.White)
		if !maximizingPlayer {
			isCheckI = move.IsKingInCheck(newBoardI, board.Black)
			isCheckJ = move.IsKingInCheck(newBoardJ, board.Black)
		}

		// Приоритеты
		scoreI := 0
		if isCheckI {
			scoreI += 100
		}
		if targetPieceI != board.Empty {
			scoreI += 10
		}
		if pieceI == board.Pawn && (moveI.ToX == 0 || moveI.ToX == 7) {
			scoreI += 5
		}
		// Добавляем небольшой случайный фактор для разнообразия
		// scoreI += rand.Intn(3)

		scoreJ := 0
		if isCheckJ {
			scoreJ += 100
		}
		if targetPieceJ != board.Empty {
			scoreJ += 10
		}
		if pieceJ == board.Pawn && (moveJ.ToX == 0 || moveJ.ToX == 7) {
			scoreJ += 5
		}
		// scoreJ += rand.Intn(3)

		return scoreI > scoreJ
	})
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
