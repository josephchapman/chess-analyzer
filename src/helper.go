package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
)

// queries a URL and returns the data as a Golang map
func queryAPI(url string) (data map[string]interface{}, err error) {

	// Get all data from URL
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("http.Get: %w", err)
		return nil, WrapError(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("resp.StatusCode: %d", resp.StatusCode)
		return nil, WrapError(err)
	}

	// Get the body of the response from the ReaderCloser interface into a Go variable 'body'
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("io.ReadAll: %w", err)
		return nil, WrapError(err)
	}

	// Convert the JSON data within 'body' to a Golang map in the 'data' var
	err = json.Unmarshal(body, &data)
	if err != nil {
		err = fmt.Errorf("json.Unmarshal: %w", err)
		return nil, WrapError(err)
	}

	return data, WrapError(err)
}

// archiveData + uuid to result
func createResultFromArchiveDataAndUUID(ad archiveData, uuid string) (r result, err error) {
	games, ok := ad.ArchiveData[ad.Key].([]interface{})
	if !ok {
		err = fmt.Errorf("key not found or is not a []interface{}")
		return result{}, WrapError(err)
	}

	for _, game := range games {
		// assert game from an interface to a map and check if the UUID matches
		gameMap, ok := game.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("game is not a map[string]interface{}")
			return result{}, WrapError(err)
		}
		if gameMap["uuid"] == uuid {
			// when the game matching the UUID is found, extract the required data
			r, err := NewResult(gameMap)
			if err != nil {
				err = fmt.Errorf("NewResult: %w", err)
				return result{}, WrapError(err)
			}
			return r, nil
		}
	}

	err = fmt.Errorf("game with UUID %s not found", uuid)
	return result{}, WrapError(err)
	// todo: better error handling in case uuid not in archiveData
}

func moveHistoryToAnalysis(mh []*chess.MoveHistory) (a analysis, err error) {
	moves := make(map[string]interface{})

	turnIncrement := 0
	whiteBestMoveHit := 0
	whiteBestMoveMiss := 0
	blackBestMoveHit := 0
	blackBestMoveMiss := 0
	for _, mh := range mh {
		// for each move
		bestMove, err := bestMoveFromPosition(mh.PrePosition)
		if err != nil {
			err = fmt.Errorf("bestMoveFromPosition: %w", err)
			return analysis{}, WrapError(err)
		}
		bestMoveAlgebraic := chess.AlgebraicNotation{}.Encode(mh.PrePosition, bestMove)
		bestMovePost := mh.PrePosition.Update(bestMove)
		bestMovePostFEN := bestMovePost.String()

		var turnString string
		if mh.PrePosition.Turn() == chess.White {
			// if it's white's turn, use a single dot in the turnString and increment white's hit/miss counters
			turnIncrement++
			turnString = fmt.Sprintf("%02d.", turnIncrement)

			if mh.PostPosition.String() == bestMovePost.String() {
				// if actual position after the move equals best position after the move
				whiteBestMoveHit++
				fmt.Println("White HIT the best move. Total:", whiteBestMoveHit)
			} else {
				whiteBestMoveMiss++
				fmt.Println("White MISSED the best move. Total:", whiteBestMoveMiss)
			}

			fmt.Println(turnString, "  (White)")
		} else {
			// if it's black's turn, use three dots in the turnString and increment black's hit/miss counters
			turnString = fmt.Sprintf("%02d...", turnIncrement)

			if mh.PostPosition == bestMovePost {
				// if actual position after the move equals best position after the move
				blackBestMoveHit++
				fmt.Println("Black HIT the best move. Total:", blackBestMoveHit)
			} else {
				blackBestMoveMiss++
				fmt.Println("Black MISSED the best move. Total:", blackBestMoveMiss)
			}

			fmt.Println(turnString, "(Black)")
		}

		moves[turnString] = make(map[string]interface{})
		moves[turnString].(map[string]interface{})["pre"] = mh.PrePosition.String()
		moves[turnString].(map[string]interface{})["actual"] = make(map[string]string)
		moves[turnString].(map[string]interface{})["actual"].(map[string]string)["move"] = chess.AlgebraicNotation{}.Encode(mh.PrePosition, mh.Move)
		moves[turnString].(map[string]interface{})["actual"].(map[string]string)["post"] = mh.PostPosition.String()
		moves[turnString].(map[string]interface{})["best"] = make(map[string]string)
		moves[turnString].(map[string]interface{})["best"].(map[string]string)["move"] = bestMoveAlgebraic
		moves[turnString].(map[string]interface{})["best"].(map[string]string)["post"] = bestMovePostFEN
	}

	whiteAccuracy := float64(whiteBestMoveHit) / float64(whiteBestMoveHit+whiteBestMoveMiss)
	blackAccuracy := float64(blackBestMoveHit) / float64(blackBestMoveHit+blackBestMoveMiss)

	a = analysis{
		Moves:         moves,
		WhiteAccuracy: whiteAccuracy,
		BlackAccuracy: blackAccuracy,
	}

	return a, nil
}

func pgnToMoveHistory(pgn string) (moveHistory []*chess.MoveHistory, err error) {
	chesspgn, err := chess.PGN(strings.NewReader(pgn))
	if err != nil {
		err = fmt.Errorf("chess.PGN: %w", err)
		return nil, WrapError(err)
	}

	game := chess.NewGame(chesspgn)
	moveHistory = game.MoveHistory()
	return moveHistory, nil
}

func bestMoveFromPosition(pos *chess.Position) (move *chess.Move, err error) {
	eng, err := uci.New("stockfish")
	if err != nil {
		err = fmt.Errorf("uci.New: %w", err)
		return nil, WrapError(err)
		// panic(err)
	}

	defer eng.Close()

	// initialize uci with new game
	err = eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame)
	if err != nil {
		err = fmt.Errorf("eng.Run: %w", err)
		return nil, WrapError(err)
		// panic(err)
	}

	cmdPos := uci.CmdPosition{Position: pos}
	// cmdGo := uci.CmdGo{Depth: 5}
	cmdGo := uci.CmdGo{MoveTime: time.Second * 1}

	err = eng.Run(cmdPos, cmdGo)
	if err != nil {
		err = fmt.Errorf("eng.Run: %w", err)
		return nil, WrapError(err)
		// panic(err)
	}

	move = eng.SearchResults().BestMove

	return move, nil
}

func epochToTime(epoch float64) time.Time {
	// Convert float64 to int64 by truncating the decimal part
	seconds := int64(epoch)

	// Convert epoch time to time.Time object
	t := time.Unix(seconds, 0)
	// fmt.Println("Converted time:", t)
	return t
}

func extractFromArchiveURL(archiveDataUrl string) (player string, yearStr string, monthStr string, err error) {
	// expected format of archiveDataUrl string:
	//   https://api.chess.com/pub/player/PLAYERNAME/games/2025/02
	//                                    ^^^^^^^^^^       ^^^^ ^^
	//                                    player           year month
	//   ^^^^^   ^^^^^^^^^^^^^ ^^^ ^^^^^^ ^^^^^^^^^^ ^^^^^ ^^^^ ^^
	//   0       2             3   4      5          6     7    8
	parts := strings.Split(archiveDataUrl, "/")

	player = parts[5]
	yearStr = parts[7]
	monthStr = parts[8]

	// Check if yearStr is a four-digit string
	if len(yearStr) != 4 || !isDigitsOnly(yearStr) {
		err = fmt.Errorf("invalid year format: %s", yearStr)
		return "", "", "", WrapError(err)
	}

	// Check if monthStr is a two-digit string
	if len(monthStr) != 2 || !isDigitsOnly(monthStr) {
		err = fmt.Errorf("invalid month format: %s", monthStr)
		return "", "", "", WrapError(err)
	}

	return player, yearStr, monthStr, nil
}

func archiveToYearMonth(archive string) (year int, month time.Month, err error) {
	// Split the archive string to get year and month
	parts := strings.Split(archive, "-")
	if len(parts) != 2 {
		fmt.Println("Error: invalid archive format. Expected format 'YYYY-MM'.")
		return 0, 0, WrapError(err)
	}

	// Parse year
	year, err = strconv.Atoi(parts[0])
	if err != nil {
		err = fmt.Errorf("strconv.Atoi: %w", err)
		return 0, 0, WrapError(err)
	}

	// Parse month
	monthInt, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Println("Error parsing month:", err)
		return
	}
	month = time.Month(monthInt)

	return year, month, WrapError(err)
}

// Helper function to check if a string contains only digits
func isDigitsOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
