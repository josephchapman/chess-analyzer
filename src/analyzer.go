package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/notnil/chess"
)

// analysis struct
type analysis struct {
	Moves         map[string]interface{} `json:"moves"`
	WhiteAccuracy float64                `json:"white_accuracy"`
	BlackAccuracy float64                `json:"black_accuracy"`
}

type player struct {
	UUID     string  `json:"uuid"`
	Username string  `json:"username"`
	Rating   float64 `json:"rating"`
}

// player creator function
func NewPlayer(uuid string, username string, rating float64) (p player) {
	p = player{
		UUID:     uuid,
		Username: username,
		Rating:   rating,
	}
	return p
}

type result struct {
	UUID        string               `json:"uuid"`
	Date        time.Time            `json:"date"`
	TimeClass   string               `json:"time_class"`
	TimeControl string               `json:"time_control"`
	Rated       bool                 `json:"rated"`
	URL         string               `json:"url"`
	PlayerWhite player               `json:"player_white"`
	PlayerBlack player               `json:"player_black"`
	PGN         string               `json:"pgn"`
	MoveHistory []*chess.MoveHistory `json:"move_history"`
	MoveCount   int                  `json:"move_count"`
	Winner      chess.Color          `json:"winner"` // White / Black / NoColor(draw)
	Outcome     string               `json:"outcome"`
	Analysis    analysis             `json:"analysis"`
}

// result creator function
func NewResult(gameData map[string]interface{}) (r result, err error) {
	whiteMap := gameData["white"].(map[string]interface{})
	playerWhiteUUID := whiteMap["uuid"].(string)
	playerWhiteUsername := whiteMap["username"].(string)
	playerWhiteRating := whiteMap["rating"].(float64)

	blackMap := gameData["black"].(map[string]interface{})
	playerBlackUUID := blackMap["uuid"].(string)
	playerBlackUsername := blackMap["username"].(string)
	playerBlackRating := blackMap["rating"].(float64)

	var outcome string
	var winner chess.Color
	if whiteMap["result"] == "win" {
		winner = chess.White
		outcome = blackMap["result"].(string)
	} else if blackMap["result"] == "win" {
		winner = chess.Black
		outcome = whiteMap["result"].(string)
	} else {
		winner = chess.NoColor
		outcome = whiteMap["result"].(string)
	}

	moveHistory, err := pgnToMoveHistory(gameData["pgn"].(string))
	if err != nil {
		err = fmt.Errorf("pgnToMoveHistory: %w", err)
		return result{}, WrapError(err)
	}

	// calculate moveCount from plies (half-moves)
	plies := len(moveHistory)
	moveCount := 0
	if plies%2 == 0 {
		moveCount = plies / 2
	} else {
		moveCount = (plies + 1) / 2
	}

	r = result{
		UUID:        gameData["uuid"].(string),
		Date:        epochToTime(gameData["end_time"].(float64)),
		TimeClass:   gameData["time_class"].(string),
		TimeControl: gameData["time_control"].(string),
		Rated:       gameData["rated"].(bool),
		URL:         gameData["url"].(string),
		PlayerWhite: NewPlayer(playerWhiteUUID, playerWhiteUsername, playerWhiteRating),
		PlayerBlack: NewPlayer(playerBlackUUID, playerBlackUsername, playerBlackRating),
		PGN:         gameData["pgn"].(string),
		MoveHistory: moveHistory,
		MoveCount:   moveCount,
		Winner:      winner,
		Outcome:     outcome,
		Analysis:    analysis{}, // populated by newAnalysis() on next line
	}

	// search the Analysis database for an existing analysis
	hasAnalysis, existingAnalysis, err := r.hasAnalysis()
	if err != nil {
		err = fmt.Errorf("r.hasAnalysis: %w", err)
		return result{}, WrapError(err)
	}
	if hasAnalysis {
		movesMap := existingAnalysis.(map[string]interface{})["moves"].(map[string]interface{})
		accuracyMap := existingAnalysis.(map[string]interface{})["accuracy"].(map[string]interface{})
		// if existing analysis, populate the object with its data
		r.Analysis = analysis{
			Moves:         movesMap,
			WhiteAccuracy: accuracyMap["white"].(float64),
			BlackAccuracy: accuracyMap["black"].(float64),
		}
	}

	return r, nil
}

// populate the Analysis field in the result object
func (r *result) analyzeGame() (err error) {
	hasAnalysis, _, err := r.hasAnalysis()
	if err != nil {
		err = fmt.Errorf("r.hasAnalysis: %w", err)
		return WrapError(err)
	}
	if hasAnalysis {
		fmt.Println("Existing analysis found in database. Not analyzing again.")
	} else {
		fmt.Println("Analyzing game:", r.UUID, "...")

		r.Analysis, err = moveHistoryToAnalysis(r.MoveHistory)
		if err != nil {
			err = fmt.Errorf("moveHistoryToAnalysis: %w", err)
			return WrapError(err)
		}

		analysisMap := make(map[string]interface{})
		analysisMap[r.UUID] = make(map[string]interface{})
		analysisMap[r.UUID].(map[string]interface{})["moves"] = r.Analysis.Moves
		accuracyMap := make(map[string]float64)
		accuracyMap["white"] = r.Analysis.WhiteAccuracy
		accuracyMap["black"] = r.Analysis.BlackAccuracy
		analysisMap[r.UUID].(map[string]interface{})["accuracy"] = accuracyMap

		// create a database object
		db, err := newDatabase("analysis", "")
		if err != nil {
			err = fmt.Errorf("newDatabase: %w", err)
			return WrapError(err)
		}

		// call its write method
		err = db.writeData(analysisMap)
		if err != nil {
			err = fmt.Errorf("db.writeData: %w", err)
			return WrapError(err)
		}
	}

	return nil
}

func (r *result) hasAnalysis() (hasAnalysis bool, existingAnalysis interface{}, err error) {
	db, err := newDatabase("analysis", "")
	if err != nil {
		err = fmt.Errorf("newDatabase: %w", err)
		return false, nil, WrapError(err)
	}

	existingAnalyses := db.Data
	existingAnalysis, ok := existingAnalyses[r.UUID]
	if ok {
		return true, existingAnalysis, nil
	} else {
		return false, nil, nil
	}
}

func (r *result) prettyPrint(verbosity string) (s string, err error) {
	if verbosity == "partial" {
		type partialResult struct {
			UUID      string    `json:"uuid"`
			Date      time.Time `json:"date"`
			MoveCount int       `json:"move_count"`
		}

		// Populate the partialResult struct with the relevant fields from the result struct
		pr := partialResult{
			UUID:      r.UUID,
			Date:      r.Date,
			MoveCount: r.MoveCount,
		}

		// Convert the partialResult struct to pretty-printed JSON
		resultJSON, err := json.MarshalIndent(pr, "", "  ")
		if err != nil {
			fmt.Println("json.MarshalIndent: %w", err)
			return "", WrapError(err)
		}

		// Print the pretty-printed JSON
		s = string(resultJSON)

	} else if verbosity == "full" {
		// Convert the result struct to pretty-printed JSON
		resultJSON, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			fmt.Println("json.MarshalIndent: %w", err)
			return "", WrapError(err)
		}

		// Print the pretty-printed JSON
		s = string(resultJSON)

	} else {
		err = fmt.Errorf("verbosity required")
		return "", WrapError(err)

	}

	return s, nil
}
