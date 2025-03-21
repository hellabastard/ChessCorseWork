package search

import (
	"chess-engine/board"
	"chess-engine/evaluation"
	"chess-engine/move"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"
)

type transpositionTableStruct struct {
	sync.Mutex
	data map[string]SearchResult
}

var transpositionTable = transpositionTableStruct{
	data: make(map[string]SearchResult),
}
var killerMoves [32][2]move.Move
var history [12][64]int

type SearchResult struct {
	BestMoves []move.Move `json:"best_moves"`
	Score     int         `json:"score"`
}

type SearchStats struct {
	NodesEvaluated int
	SearchTime     time.Duration
}

func LoadData() {
	if data, err := os.ReadFile("transpositions.json"); err == nil {
		if err := json.Unmarshal(data, &transpositionTable.data); err != nil {
			fmt.Printf("Ошибка загрузки транспозиционной таблицы: %v\n", err)
		} else {
			fmt.Printf("Загружено %d позиций из транспозиционной таблицы\n", len(transpositionTable.data))
		}
	}

	if data, err := os.ReadFile("killers.json"); err == nil {
		if err := json.Unmarshal(data, &killerMoves); err != nil {
			fmt.Printf("Ошибка загрузки killer moves: %v\n", err)
		} else {
			fmt.Println("Загружены killer moves")
		}
	}
}

func SaveData() {
	transpositionTable.Lock()
	defer transpositionTable.Unlock()
	if data, err := json.MarshalIndent(transpositionTable.data, "", "  "); err == nil {
		if err := os.WriteFile("transpositions.json", data, 0644); err != nil {
			fmt.Printf("Ошибка сохранения транспозиционной таблицы: %v\n", err)
		} else {
			fmt.Printf("Сохранено %d позиций в транспозиционную таблицу\n", len(transpositionTable.data))
		}
	}

	if data, err := json.MarshalIndent(killerMoves, "", "  "); err == nil {
		if err := os.WriteFile("killers.json", data, 0644); err != nil {
			fmt.Printf("Ошибка сохранения killer moves: %v\n", err)
		} else {
			fmt.Println("Сохранены killer moves")
		}
	}
}

func Minimax(b board.Board, depth int, alpha int, beta int, maximizingPlayer bool, deadline time.Time, stats *SearchStats) SearchResult {
	if time.Now().After(deadline) {
		stats.NodesEvaluated++
		return SearchResult{Score: evaluation.Evaluate(b)}
	}

	hash := boardToString(b)
	transpositionTable.Lock()
	if result, ok := transpositionTable.data[hash]; ok && depth <= result.Score {
		transpositionTable.Unlock()
		return result
	}
	transpositionTable.Unlock()

	if depth == 0 {
		return SearchResult{Score: QuiescenceSearch(b, alpha, beta, maximizingPlayer, 4, deadline, stats)}
	}

	color := board.Black
	if maximizingPlayer {
		color = board.White
	}

	moves := move.GenerateMoves(b, color)
	if len(moves) == 0 {
		if maximizingPlayer && move.IsKingInCheck(b, board.White) {
			return SearchResult{Score: -1000000}
		} else if !maximizingPlayer && move.IsKingInCheck(b, board.Black) {
			return SearchResult{Score: 1000000}
		}
		fmt.Println("Пат или нет ходов для", color)
		stats.NodesEvaluated++
		return SearchResult{Score: evaluation.Evaluate(b)}
	}

	sortMoves(moves, b, depth)

	var bestMoves []move.Move
	var bestScore int
	if maximizingPlayer {
		bestScore = math.MinInt
	} else {
		bestScore = math.MaxInt
	}

	for _, m := range moves {
		newBoard := b.Copy()
		if err := move.MakeMove(&newBoard, m); err != nil {
			fmt.Printf("Ошибка в MakeMove для хода %v: %v\n", m, err)
			continue
		}
		res := Minimax(newBoard, depth-1, alpha, beta, !maximizingPlayer, deadline, stats)
		stats.NodesEvaluated++
		if maximizingPlayer {
			if res.Score > bestScore {
				bestScore = res.Score
				bestMoves = []move.Move{m}
			} else if res.Score == bestScore {
				bestMoves = append(bestMoves, m)
			}
			alpha = max(alpha, bestScore)
			if beta <= alpha {
				updateKillerAndHistory(b, m, depth, color)
				break
			}
		} else {
			if res.Score < bestScore {
				bestScore = res.Score
				bestMoves = []move.Move{m}
			} else if res.Score == bestScore {
				bestMoves = append(bestMoves, m)
			}
			beta = min(beta, bestScore)
			if beta <= alpha {
				updateKillerAndHistory(b, m, depth, color)
				break
			}
		}
	}

	if len(bestMoves) == 0 {
		fmt.Println("Не удалось найти лучшие ходы для", color)
		return SearchResult{Score: evaluation.Evaluate(b)}
	}

	res := SearchResult{
		BestMoves: bestMoves,
		Score:     bestScore,
	}
	transpositionTable.Lock()
	transpositionTable.data[hash] = res
	transpositionTable.Unlock()
	return res
}

func QuiescenceSearch(b board.Board, alpha int, beta int, maximizingPlayer bool, maxDepth int, deadline time.Time, stats *SearchStats) int {
	if time.Now().After(deadline) || maxDepth <= 0 {
		stats.NodesEvaluated++
		return evaluation.Evaluate(b)
	}

	standPat := evaluation.Evaluate(b)
	stats.NodesEvaluated++
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
		newBoard := b.Copy()
		if err := move.MakeMove(&newBoard, m); err != nil {
			continue
		}

		if targetPiece != board.Empty || move.IsKingInCheck(newBoard, board.Black) || move.IsKingInCheck(newBoard, board.White) ||
			(piece == board.Pawn && (m.ToX == 0 || m.ToX == 7)) {
			score := QuiescenceSearch(newBoard, alpha, beta, !maximizingPlayer, maxDepth-1, deadline, stats)
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

func FindBestMove(b board.Board, depth int, boardColor board.Color) (move.Move, SearchStats) {
	rand.Seed(time.Now().UnixNano())
	maximizingPlayer := (boardColor == board.White)
	start := time.Now()
	timeLimit := 10 * time.Second
	deadline := start.Add(timeLimit)

	stats := SearchStats{}
	res := Minimax(b, depth, math.MinInt, math.MaxInt, maximizingPlayer, deadline, &stats)
	stats.SearchTime = time.Since(start)

	if len(res.BestMoves) == 0 {
		fmt.Println("Minimax вернул пустой список лучших ходов для", boardColor)
		moves := move.GenerateMoves(b, boardColor)
		if len(moves) == 0 {
			fmt.Println("GenerateMoves вернул пустой список для", boardColor)
			return move.Move{}, stats
		}
		return moves[0], stats // Возвращаем первый доступный ход
	}

	// Ограничиваем рандомизацию топ-3 ходами (или всеми, если их меньше)
	maxChoices := 3
	if len(res.BestMoves) < maxChoices {
		maxChoices = len(res.BestMoves)
	}
	// Сортируем ходы по эвристике для дебюта
	sort.Slice(res.BestMoves, func(i, j int) bool {
		return moveHeuristic(res.BestMoves[i]) > moveHeuristic(res.BestMoves[j])
	})
	// Выбираем случайный из топ-N
	return res.BestMoves[rand.Intn(maxChoices)], stats
}

// moveHeuristic добавляет приоритет центральным ходам в дебюте
func moveHeuristic(m move.Move) int {
	score := 0
	// Предпочтение центральным клеткам (d4, d5, e4, e5)
	centerSquares := map[int]bool{27: true, 28: true, 35: true, 36: true} // e4, e5, d4, d5
	toSquare := m.ToX*8 + m.ToY
	if centerSquares[toSquare] {
		score += 10
	}
	return score
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

		scoreI := 0
		if targetPieceI != board.Empty {
			targetValue := evaluation.PieceValues[targetPieceI]
			pieceValue := evaluation.PieceValues[pieceI]
			scoreI += targetValue - pieceValue/10
		}
		if pieceI == board.Pawn && (moveI.ToX == 0 || moveI.ToX == 7) {
			scoreI += 900
		}
		if pieceI == board.Knight || pieceI == board.Bishop {
			scoreI += 20
		}
		if pieceI == board.Pawn && (moveI.ToY == 3 || moveI.ToY == 4) && targetPieceI == board.Empty {
			scoreI += 20
		}
		if depth < len(killerMoves) {
			if moveI == killerMoves[depth][0] {
				scoreI += 1000
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

func updateKillerAndHistory(b board.Board, m move.Move, depth int, color board.Color) {
	if depth < len(killerMoves) {
		killerMoves[depth][1] = killerMoves[depth][0]
		killerMoves[depth][0] = m
	}
	piece, _, _ := b.GetPiece(m.FromX, m.FromY)
	pieceIndex := int(piece) + 6*int(color)
	if pieceIndex < 12 {
		history[pieceIndex][m.ToX*8+m.ToY] += depth * depth
	}
}
