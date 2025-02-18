package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// archiveList struct
type archiveList struct {
	Player      string                 `json:"player"`
	URL         string                 `json:"url"`
	ArchiveList map[string]interface{} `json:"archives"`
	Present     map[string]bool        `json:"present"`
}

// ArchiveList creator function
func NewArchiveList(player string, source string) (al archiveList, err error) {
	archiveListUrl := fmt.Sprintf("https://api.chess.com/pub/player/%s/games/archives", player)

	al = archiveList{
		Player:      player,
		URL:         archiveListUrl,
		ArchiveList: make(map[string]interface{}), // this needs to be refreshable
		//                                            it's therefore intialized blank
		//                                            and updated via
		//                                            `getArchiveList()`
		Present: make(map[string]bool), // this needs to be refreshable
		//                                 it's therefore intialized blank
		//                                 and updated via
		//                                 `getPresent()`
	}

	if source == "db" {
		al.getArchiveListFromDB()
	} else if source == "api" {
		al.getArchiveListFromAPI()
	} else {
		err = fmt.Errorf("NewArchiveList: invalid source")
		return archiveList{}, WrapError(err)
	}

	al.getPresent()

	return al, nil
}

// archiveList methods
func (al *archiveList) getArchiveListFromDB() (err error) {
	// read from the database
	db, err := newDatabase("archive_list", al.Player)
	if err != nil {
		err = fmt.Errorf("newDatabase: %w", err)
		return WrapError(err)
	}

	// write to the object
	al.ArchiveList = db.Data

	return nil
}

func (al *archiveList) getArchiveListFromAPI() (err error) {
	// read from the API and write it to the object
	al.ArchiveList, err = queryAPI(al.URL)
	if err != nil {
		err = fmt.Errorf("queryAPI: %w", err)
		return WrapError(err)
	}
	return nil
}

func (al *archiveList) getPresent() (err error) {
	// read from the database
	db, err := newDatabase("archive_data", al.Player)
	if err != nil {
		err = fmt.Errorf("newDatabase: %w", err)
		return WrapError(err)
	}

	data := make(map[string]bool)
	yearStr := ""
	monthStr := ""
	keyStr := ""

	// Check if the 'archives' key exists and is a slice of interfaces
	archives, ok := al.ArchiveList["archives"].([]interface{})
	if !ok {
		err = fmt.Errorf("'archives' key not found or is not []interface{}")
		return WrapError(err)
	}

	for _, archive := range archives {
		// Convert archive from interface{} to string
		archiveDataUrl, ok := archive.(string)
		if !ok {
			err = fmt.Errorf("archive is not a string")
			return WrapError(err)
		}

		_, yearStr, monthStr, err = extractFromArchiveURL(archiveDataUrl)
		if err != nil {
			err = fmt.Errorf("extractFromArchiveURL: %w", err)
			return WrapError(err)
		}
		keyStr = fmt.Sprintf("%s-%s", yearStr, monthStr)

		// Check if keyStr is a key within the db.Data map
		_, exists := db.Data[keyStr]
		if exists {
			data[keyStr] = true
		} else {
			data[keyStr] = false
		}
	}

	// Sort the keys of the data map alphanumerically
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Create a new sorted map
	sortedData := make(map[string]bool)
	for _, key := range keys {
		sortedData[key] = data[key]
	}

	al.Present = data

	return nil
}

func (al *archiveList) prettyPrint() (s string, err error) {
	// Pretty-print the Present list as indented JSON
	presentJSON, err := json.MarshalIndent(al.Present, "", "  ")
	if err != nil {
		err = fmt.Errorf("json.MarshalIndent: %w", err)
		return "", WrapError(err)
	}

	s = string(presentJSON)
	return s, nil
}

// archiveData struct
type archiveData struct {
	Player      string                 `json:"player"`
	Year        int                    `json:"year"`
	Month       time.Month             `json:"month"`
	Key         string                 `json:"key"`
	URL         string                 `json:"url"`
	ArchiveData map[string]interface{} `json:"games"`
	Present     map[string]bool        `json:"present"`
}

// archiveData creator function
func NewArchiveData(player string, year int, month time.Month, source string) (archiveData, error) {
	// convert month from time.Month to int
	monthInt := int(month)

	// construct archiveMonth(string) (e.g. "2023/07") from year(int) and month(int)
	// this is to be used as the top-level key, replacing the general 'games'
	key := fmt.Sprintf("%d-%02d", year, monthInt)

	// monthInt can now be used to construct the archiveDataUrl
	archiveDataUrl := fmt.Sprintf("https://api.chess.com/pub/player/%s/games/%d/%02d", player, year, monthInt)

	ad := archiveData{
		Player:      player,
		Year:        year,
		Month:       month,
		Key:         key,
		URL:         archiveDataUrl,
		ArchiveData: make(map[string]interface{}), // this needs to be refreshable
		//                                            it's therefore intialized blank
		//                                            and updated via
		//                                            `getArchiveData()`
		Present: make(map[string]bool), // this needs to be refreshable
		//                                 it's therefore intialized blank
		//                                 and updated via
		//                                 `getPresent()`
	}
	var err error
	if source == "db" {
		ad.getArchiveDataFromDB()
	} else if source == "api" {
		err = ad.getArchiveDataFromAPI()
		if err != nil {
			err = fmt.Errorf("getArchiveDataFromAPI: %w", err)
			return archiveData{}, WrapError(err)
		}
	} else {
		err = fmt.Errorf("invalid 'source' variable passed to ")
		return archiveData{}, WrapError(err)
	}

	ad.getPresent()

	return ad, WrapError(err)
}

// archiveData methods
func (ad *archiveData) getArchiveDataFromDB() (err error) {
	// read from the database
	db, err := newDatabase("archive_data", ad.Player)
	if err != nil {
		err = fmt.Errorf("newDatabase: %w", err)
		return WrapError(err)
	}

	data := make(map[string]interface{})
	data[ad.Key] = db.Data[ad.Key]
	ad.ArchiveData = data

	return nil
}

// Populate the ArchiveData field via an API call
func (ad *archiveData) getArchiveDataFromAPI() (err error) {

	// read from the API and write it to the object
	apiData, err := queryAPI(ad.URL)
	if err != nil {
		err = fmt.Errorf("queryAPI: %w", err)
		return WrapError(err)
	}
	data := make(map[string]interface{})
	data[ad.Key] = apiData["games"]
	ad.ArchiveData = data

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

	return nil
}

func (ad *archiveData) getPresent() (err error) {
	// read from the _analysis database
	db, err := newDatabase("analysis", "")
	if err != nil {
		err = fmt.Errorf("newDatabase: %w", err)
		return WrapError(err)
	}

	data := make(map[string]bool)
	uuid := ""

	// games should be a list of interfaces
	games, ok := ad.ArchiveData[ad.Key].([]interface{})
	if !ok {
		err = fmt.Errorf("key not found or is not []interface{}")
		return WrapError(err)
	}

	for _, game := range games {
		gameMap, ok := game.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("game is not a map[string]interface{}")
			return WrapError(err)
		}

		uuid = gameMap["uuid"].(string)

		// Check if uuid is a key within the db.Data map
		if _, exists := db.Data[uuid]; exists {
			data[uuid] = true
		} else {
			data[uuid] = false
		}
	}

	ad.Present = data

	return nil
}

func (ad *archiveData) prettyPrint() (s string, err error) {
	// Pretty-print the Present list as indented JSON
	presentJSON, err := json.MarshalIndent(ad.Present, "", "  ")
	if err != nil {
		err = fmt.Errorf("json.MarshalIndent: %w", err)
		return "", WrapError(err)
	}

	s = string(presentJSON)
	return s, nil
}
