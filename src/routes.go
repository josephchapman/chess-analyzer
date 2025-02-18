package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// https://go.dev/blog/error-handling-and-go
// Handles errors and logging
type appHandler func(http.ResponseWriter, *http.Request) (err error)

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Logging the request
	log.SetOutput(os.Stdout)
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	// Handling the request and capturing any error
	err := fn(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GET /healthz
func Healthz(w http.ResponseWriter, r *http.Request) (err error) {
	data := "OK"
	// Write the JSON response
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(data))

	return nil
}

// GET /api/{player}
func APIarchiveListGet(w http.ResponseWriter, r *http.Request) (err error) {
	player := r.PathValue("player")

	al, err := NewArchiveList(player, "db")
	if err != nil {
		err = fmt.Errorf("NewArchiveList: %w", err)
		return WrapError(err)
	}

	data, err := al.prettyPrint()
	if err != nil {
		err = fmt.Errorf("al.prettyPrint: %w", err)
		return WrapError(err)
	}
	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))

	return nil
}

// POST /api/{player}
func APIarchiveListPost(w http.ResponseWriter, r *http.Request) (err error) {
	player := r.PathValue("player")

	al, err := NewArchiveList(player, "api")
	if err != nil {
		err = fmt.Errorf("NewArchiveList: %w", err)
		return WrapError(err)
	}

	// read from the object and write it to the database
	db, err := newDatabase("archive_list", al.Player)
	if err != nil {
		err = fmt.Errorf("newDatabase: %w", err)
		return WrapError(err)
	}

	err = db.writeData(al.ArchiveList)
	if err != nil {
		err = fmt.Errorf("db.writeData: %w", err)
		return WrapError(err)
	}

	// Write the plaintext response
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Archive List Updated"))

	return nil
}

// GET /api/{player}/{archive}
func APIarchiveDataGet(w http.ResponseWriter, r *http.Request) (err error) {
	player := r.PathValue("player")
	archive := r.PathValue("archive")

	year, month, err := archiveToYearMonth(archive)
	if err != nil {
		err = fmt.Errorf("archiveToYearMonth: %w", err)
		return WrapError(err)
	}

	ad, err := NewArchiveData(player, year, month, "db")
	if err != nil {
		err = fmt.Errorf("NewArchiveData: %w", err)
		return WrapError(err)
	}

	data, err := ad.prettyPrint()
	if err != nil {
		err = fmt.Errorf("ad.prettyPrint: %w", err)
		return WrapError(err)
	}
	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))

	return nil
}

// POST /api/{player}/{archive}
func APIarchiveDataPost(w http.ResponseWriter, r *http.Request) (err error) {
	player := r.PathValue("player")
	archive := r.PathValue("archive")

	year, month, err := archiveToYearMonth(archive)
	if err != nil {
		err = fmt.Errorf("archiveToYearMonth: %w", err)
		return WrapError(err)
	}

	ad, err := NewArchiveData(player, year, month, "api")
	if err != nil {
		err = fmt.Errorf("NewArchiveData: %w", err)
		return WrapError(err)
	}

	// read from the object and write it to the database
	db, err := newDatabase("archive_data", ad.Player)
	if err != nil {
		err = fmt.Errorf("newDatabase: %w", err)
		return WrapError(err)
	}

	err = db.writeData(ad.ArchiveData)
	if err != nil {
		err = fmt.Errorf("db.writeData: %w", err)
		return WrapError(err)
	}

	// Write the plaintext response
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Archive Data Updated"))

	return nil
}

// GET /api/{player}/{archive}/{uuid}
func APIresultGet(w http.ResponseWriter, r *http.Request) (err error) {
	player := r.PathValue("player")
	archive := r.PathValue("archive")
	uuid := r.PathValue("uuid")
	verbosity := "full"

	year, month, err := archiveToYearMonth(archive)
	if err != nil {
		err = fmt.Errorf("archiveToYearMonth: %w", err)
		return WrapError(err)
	}

	ad, err := NewArchiveData(player, year, month, "db")
	if err != nil {
		err = fmt.Errorf("NewArchiveData: %w", err)
		return WrapError(err)
	}

	result, err := createResultFromArchiveDataAndUUID(ad, uuid)
	if err != nil {
		err = fmt.Errorf("createResultFromArchiveDataAndUUID: %w", err)
		return WrapError(err)
	}
	data, err := result.prettyPrint(verbosity)
	if err != nil {
		err = fmt.Errorf("result.prettyPrint: %w", err)
		return WrapError(err)
	}
	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))

	return nil
}

// POST /api/{player}/{archive}/{uuid}
func APIresultPost(w http.ResponseWriter, r *http.Request) (err error) {
	player := r.PathValue("player")
	archive := r.PathValue("archive")
	uuid := r.PathValue("uuid")

	year, month, err := archiveToYearMonth(archive)
	if err != nil {
		err = fmt.Errorf("archiveToYearMonth: %w", err)
		return WrapError(err)
	}

	ad, err := NewArchiveData(player, year, month, "db")
	if err != nil {
		err = fmt.Errorf("NewArchiveData: %w", err)
		return WrapError(err)
	}

	result, err := createResultFromArchiveDataAndUUID(ad, uuid)
	if err != nil {
		err = fmt.Errorf("createResultFromArchiveDataAndUUID: %w", err)
		return WrapError(err)
	}

	result.analyzeGame()

	// Write the plaintext response
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Result Updated"))

	return nil
}
