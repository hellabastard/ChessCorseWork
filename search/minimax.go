package search

import (
	"chess-engine/board"
	"chess-engine/evaluation"
	"chess-engine/move"
	"fmt"
	"math"
	"sort"
	"time"
)

var transpositionTable = make(map[string]SearchResult)
var killerMoves [32][2]move.Move
var history [12][64]int

type SearchResult struct {
	BestMove move.Move
	Score    int
}

func Minimax(b board.Board, depth int, alpha int, beta int, maximizingPlayer bool, deadline time.Time) SearchResult {
	if time.Now().After(deadline) {
		return SearchResult{Score: evaluation.Evaluate(b)}
	}

	hash := boardToString(b)
	if result, ok := transpositionTable[hash]; ok && depth <= result.Score {
		return result
	}

	if depth == 0 {
		return SearchResult{Score: QuiescenceSearch(b, alpha, beta, maximizingPlayer, 4, deadline)}
	}

	color := board.Black
	if maximizingPlayer {
		color = board.White
	}
	if depth > 2 && !move.IsKingInCheck(b, color) {
		nullBoard := b
		nullResult := Minimax(nullBoard, depth-3, alpha, beta, !maximizingPlayer, deadline)
		if maximizingPlayer && nullResult.Score >= beta {
			return SearchResult{Score: beta}
		} else if !maximizingPlayer && nullResult.Score <= alpha {
			return SearchResult{Score: alpha}
		}
	}

	var bestMove move.Move
	var bestScore int
	if maximizingPlayer {
		bestScore = math.MinInt
	} else {
		bestScore = math.MaxInt
	}

	moves := move.GenerateMoves(b, color)
	if len(moves) == 0 {
		if maximizingPlayer && move.IsKingInCheck(b, board.White) {
			return SearchResult{Score: -1000000}
		} else if !maximizingPlayer && move.IsKingInCheck(b, board.Black) {
			return SearchResult{Score: 1000000}
		}
		return SearchResult{Score: evaluation.Evaluate(b)}
	}

	sortMoves(moves, b, depth)

	for _, m := range moves {
		newBoard := b
		if err := move.MakeMove(&newBoard, m); err != nil {
			continue
		}

		result := Minimax(newBoard, depth-1, alpha, beta, !maximizingPlayer, deadline)
		if maximizingPlayer {
			if result.Score > bestScore {
				bestScore = result.Score
				bestMove = m
			}
			alpha = max(alpha, bestScore)
			if beta <= alpha {
				if depth < len(killerMoves) {
					killerMoves[depth][1] = killerMoves[depth][0]
					killerMoves[depth][0] = m
				}
				piece, _, _ := b.GetPiece(m.FromX, m.FromY)
				pieceIndex := int(piece) + 6*int(color)
				if pieceIndex < 12 {
					history[pieceIndex][m.ToX*8+m.ToY] += depth * depth
				}
				break
			}
		} else {
			if result.Score < bestScore {
				bestScore = result.Score
				bestMove = m
			}
			beta = min(beta, bestScore)
			if beta <= alpha {
				if depth < len(killerMoves) {
					killerMoves[depth][1] = killerMoves[depth][0]
					killerMoves[depth][0] = m
				}
				piece, _, _ := b.GetPiece(m.FromX, m.FromY)
				pieceIndex := int(piece) + 6*int(color)
				if pieceIndex < 12 {
					history[pieceIndex][m.ToX*8+m.ToY] += depth * depth
				}
				break
			}
		}
	}

	result := SearchResult{
		BestMove: bestMove,
		Score:    bestScore,
	}
	transpositionTable[hash] = result
	return result
}

func QuiescenceSearch(b board.Board, alpha int, beta int, maximizingPlayer bool, maxDepth int, deadline time.Time) int {
	if time.Now().After(deadline) || maxDepth <= 0 {
		return evaluation.Evaluate(b)
	}

	standPat := evaluation.Evaluate(b)
	if maximizingPlayer {
		if standPat >= beta {
			return beta
		}
		alpha = max(alpha, standPat)
	} else {
		if standPat <= alpha {
			return alpha
		}
		beta = min(beta, standPat)
	}

	color := board.Black
	if maximizingPlayer {
		color = board.White
	}

	moves := move.GenerateMoves(b, color)
	sortMoves(moves, b, 0)

	for _, m := range moves {
		targetPiece, _, _ := b.GetPiece(m.ToX, m.ToY)
		piece, _, _ := b.GetPiece(m.FromX, m.FromY)
		newBoard := b
		if err := move.MakeMove(&newBoard, m); err != nil {
			continue
		}

		if targetPiece != board.Empty || move.IsKingInCheck(newBoard, board.Black) || move.IsKingInCheck(newBoard, board.White) ||
			(piece == board.Pawn && (m.ToX == 0 || m.ToX == 7)) {
			score := QuiescenceSearch(newBoard, alpha, beta, !maximizingPlayer, maxDepth-1, deadline)
			if maximizingPlayer {
				alpha = max(alpha, score)
				if alpha >= beta {
					break
				}
			} else {
				beta = min(beta, score)
				if beta <= alpha {
					break
				}
			}
		}
	}

	if maximizingPlayer {
		return alpha
	}
	return beta
}

func FindBestMove(b board.Board, depth int, boardColor board.Color) move.Move {
	maximizingPlayer := (boardColor == board.White)
	start := time.Now()
	timeLimit := 10 * time.Second // Установлено 10 секунд
	deadline := start.Add(timeLimit)

	result := Minimax(b, depth, math.MinInt, math.MaxInt, maximizingPlayer, deadline)
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

func sortMoves(moves []move.Move, b board.Board, depth int) {
	sort.Slice(moves, func(i, j int) bool {
		moveI, moveJ := moves[i], moves[j]

		if moveI.FromX < 0 || moveI.FromX >= 8 || moveI.FromY < 0 || moveI.FromY >= 8 ||
			moveI.ToX < 0 || moveI.ToX >= 8 || moveI.ToY < 0 || moveI.ToY >= 8 ||
			moveJ.FromX < 0 || moveJ.FromX >= 8 || moveJ.FromY < 0 || moveJ.FromY >= 8 ||
			moveJ.ToX < 0 || moveJ.ToX >= 8 || moveJ.ToY < 0 || moveJ.ToY >= 8 {
			return false
		}

		pieceI, colorI, _ := b.GetPiece(moveI.FromX, moveI.FromY)
		pieceJ, colorJ, _ := b.GetPiece(moveJ.FromX, moveJ.FromY)
		targetPieceI, _, _ := b.GetPiece(moveI.ToX, moveI.ToY)
		targetPieceJ, _, _ := b.GetPiece(moveJ.ToX, moveJ.ToY)

		// MVV-LVA: приоритет взятий более ценных фигур менее ценными
		scoreI := 0
		if targetPieceI != board.Empty {
			targetValue := evaluation.PieceValues[targetPieceI] // Теперь доступно
			pieceValue := evaluation.PieceValues[pieceI]
			scoreI += targetValue - pieceValue/10 // Бонус за взятие ценной фигуры
		}
		if pieceI == board.Pawn && (moveI.ToX == 0 || moveI.ToX == 7) {
			scoreI += 900 // Превращение пешки
		}
		if pieceI == board.Knight || pieceI == board.Bishop {
			scoreI += 20 // Развитие лёгких фигур
		}
		if pieceI == board.Pawn && (moveI.ToY == 3 || moveI.ToY == 4) && targetPieceI == board.Empty {
			scoreI += 20 // Центр
		}
		if depth < len(killerMoves) {
			if moveI == killerMoves[depth][0] {
				scoreI += 1000 // Killer moves имеют высокий приоритет
			} else if moveI == killerMoves[depth][1] {
				scoreI += 900
			}
		}
		pieceIndexI := int(pieceI) + 6*int(colorI)
		if pieceIndexI < 12 {
			scoreI += history[pieceIndexI][moveI.ToX*8+moveI.ToY] / 100
		}

		scoreJ := 0
		if targetPieceJ != board.Empty {
			targetValue := evaluation.PieceValues[targetPieceJ]
			pieceValue := evaluation.PieceValues[pieceJ]
			scoreJ += targetValue - pieceValue/10
		}
		if pieceJ == board.Pawn && (moveJ.ToX == 0 || moveJ.ToX == 7) {
			scoreJ += 900
		}
		if pieceJ == board.Knight || pieceJ == board.Bishop {
			scoreJ += 20
		}
		if pieceJ == board.Pawn && (moveJ.ToY == 3 || moveJ.ToY == 4) && targetPieceJ == board.Empty {
			scoreJ += 20
		}
		if depth < len(killerMoves) {
			if moveJ == killerMoves[depth][0] {
				scoreJ += 1000
			} else if moveJ == killerMoves[depth][1] {
				scoreJ += 900
			}
		}
		pieceIndexJ := int(pieceJ) + 6*int(colorJ)
		if pieceIndexJ < 12 {
			scoreJ += history[pieceIndexJ][moveJ.ToX*8+moveJ.ToY] / 100
		}

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
