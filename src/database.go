package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
)

// database struct
type database struct {
	TableName   string                 `json:"table_name"`
	ContentType string                 `json:"contents"` // list or data
	Data        map[string]interface{} `json:"data"`
}

// database creator function
func newDatabase(contentType string, player string) (db database, err error) {
	tableName := fmt.Sprintf("%s_%s.json", player, contentType)

	db = database{
		TableName:   tableName,
		ContentType: contentType,
		Data:        make(map[string]interface{}), // this needs to be refreshable
		//                                            it's therefore intialized blank
		//                                            and updated via
		//                                            `readData()`
	}

	err = db.readData()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// If the file does not exist
			fmt.Println(err)
		}
		// do not return an error
	}

	return db, nil
}

func (db *database) getFilePath() string {
	return filepath.Join("/var/lib/data", db.TableName)
}

func (db *database) readData() (err error) {
	// Initialize the data map
	data := make(map[string]interface{})

	// Open the file
	file, err := os.Open(db.getFilePath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// If the file does not exist
			err = fmt.Errorf("file does not exist: %w", err)
			return WrapError(err)
		}
		err = fmt.Errorf("os.Open: %w", err)
		return WrapError(err)
	}

	defer file.Close()

	// Decode the JSON content
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		err = fmt.Errorf("decoder.Decode: %w", err)
		return WrapError(err)
	}

	db.Data = data

	return nil
}

func (db *database) writeData(data map[string]interface{}) (err error) {
	err = db.readData() // refresh the data in the database object
	if err != nil {
		err = fmt.Errorf("db.readData: %w", err)
		// do not return an error
	}

	existingData := db.Data // read the data from the database object
	maps.Copy(existingData, data)

	// Create the file on disk
	file, err := os.Create(db.getFilePath())
	if err != nil {
		err = fmt.Errorf("os.Create: %w", err)
		return WrapError(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(existingData)
	if err != nil {
		err = fmt.Errorf("encoder.Encode: %w", err)
		return WrapError(err)
	}

	return nil
}
