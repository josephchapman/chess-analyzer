package main

import (
	"log"
	"os"
)

// Custom error type that wraps the original error and logs it
type LoggedError struct {
	Err error
}

func (e *LoggedError) Error() string {
	return e.Err.Error()
}

// WrapError logs the error and returns a LoggedError
func WrapError(err error) error {
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Println(err)
		return &LoggedError{Err: err}
	}
	return nil
}
