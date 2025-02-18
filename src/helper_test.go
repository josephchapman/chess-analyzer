package main

import (
	"testing"
	"time"
)

func TestExtractFromArchiveURL(t *testing.T) {
	type testCase struct {
		// Input Params
		archiveDataUrl string
		// Expected Values
		player   string
		yearStr  string
		monthStr string
		err      error
	}

	t.Run("is digits only", func(t *testing.T) {
		tests := []testCase{
			{"https://api.chess.com/pub/player/asdf/games/2020/08", "asdf", "2020", "08", nil},
			{"https://api.chess.com/pub/player/1234/games/2021/09", "1234", "2021", "09", nil},
			{"https://api.chess.com/pub/player/4f5g/games/2022/10", "4f5g", "2022", "10", nil},
		}

		for _, test := range tests {
			actualPlayer, actualYearStr, actualMonthStr, actualErr := extractFromArchiveURL(test.archiveDataUrl)
			if actualPlayer != test.player {
				t.Errorf("expected %v, got %v", test.player, actualPlayer)
			}
			if actualYearStr != test.yearStr {
				t.Errorf("expected %v, got %v", test.yearStr, actualYearStr)
			}
			if actualMonthStr != test.monthStr {
				t.Errorf("expected %v, got %v", test.monthStr, actualMonthStr)
			}
			if actualErr != test.err {
				t.Errorf("expected %v, got %v", test.err, actualErr)
			}
		}
	})
}

func TestArchiveToYearMonth(t *testing.T) {
	type testCase struct {
		// Input Params
		archive string
		// Expected Values
		year  int
		month time.Month
		err   error
	}

	t.Run("archive to year and month", func(t *testing.T) {
		tests := []testCase{
			{"2020-05", 2020, 5, nil},
			{"2021-06", 2021, 6, nil},
		}

		for _, test := range tests {
			actualYear, actualMonth, actualErr := archiveToYearMonth(test.archive)
			if actualYear != test.year {
				t.Errorf("expected %v, got %v", test.year, actualYear)
			}
			if actualMonth != test.month {
				t.Errorf("expected %v, got %v", test.month, actualMonth)
			}
			if actualErr != test.err {
				t.Errorf("expected %v, got %v", test.err, actualErr)
			}
		}
	})
}

func TestIsDigitsOnly(t *testing.T) {
	type testCase struct {
		// Input Params
		s string
		// Expected Values
		b bool
	}

	t.Run("is digits only", func(t *testing.T) {
		tests := []testCase{
			{"1234567890", true},
			{"asdfghjkl", false},
			{"12as34df56gh", false},
			{"+)_", false},
			{"tuVijmuiwz7xYsni", false},
		}

		for _, test := range tests {
			actual := isDigitsOnly(test.s)
			if actual != test.b {
				t.Errorf("expected %v, got %v", test.b, actual)
			}
		}
	})
}
