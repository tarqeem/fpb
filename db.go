package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

const db_path string = "fbp.db"

var Db *sql.DB

// initDb creates the main table if it doesn't exist
func initDb() (err error) {
	Db, err = sql.Open("sqlite3", db_path)
	if err != nil {
		return err
	}

	// Create main table
	createTableSQL := `
        CREATE TABLE IF NOT EXISTS main_table (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            json TEXT UNIQUE NOT NULL
        );
    `
	_, err = Db.Exec(createTableSQL)
	return err

}

// AddJSONToDB adds a JSON string to the database with today's date if it
// doesn't already exist
func AddJSONToDB(db *sql.DB, jsonStr string) error {
	// Check if the JSON already exists
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM main_table WHERE json = ?", jsonStr)
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		log.Info().Msg("JSON already exists in the database")
		return nil
	}

	// Insert JSON into the database with today's date
	today := time.Now().Format(TimeFormat)
	_, err := db.Exec("INSERT INTO main_table (date, json) VALUES (?, ?)", today, jsonStr)
	if err != nil {
		return err
	}

	return nil
}

func getJSONFromDB(db *sql.DB, getAll bool) (cookies []Cookie, err error) {
	today := time.Now().Format(TimeFormat)

	// Retrieve JSON strings for today or older than today
	var rows *sql.Rows
	if getAll {
		rows, err = db.Query("SELECT json FROM main_table WHERE date <= ?", today)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else {
		row := db.QueryRow("SELECT json FROM main_table WHERE date <= ? LIMIT 1", today)
		var jsonStr string
		if err := row.Scan(&jsonStr); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil // No cookie found
			}
			return nil, err
		}
		rows = &sql.Rows{}
		rows.Close()
		rows, err = db.Query("SELECT json FROM main_table WHERE json = ?", jsonStr)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	}

	var jsonList []string
	for rows.Next() {
		var jsonStr string
		if err := rows.Scan(&jsonStr); err != nil {
			return nil, err
		}
		jsonList = append(jsonList, jsonStr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, jsonRow := range jsonList {
		var cookie Cookie
		if err := json.Unmarshal([]byte(jsonRow), &cookie); err != nil {
			return nil, err
		}
		cookies = append(cookies, cookie)
	}

	return cookies, nil
}

func DeleteJSONFromDB(db *sql.DB, jsonStr string) error {
	// Delete the JSON from the database
	_, err := db.Exec("DELETE FROM main_table WHERE json = ?", jsonStr)
	return err
}

func IncreaseDateByNDays(db *sql.DB, jsonStr string, n int) error {
	// Get the current date
	rows, err := db.Query("SELECT date FROM main_table WHERE json = ?", jsonStr)
	if err != nil {
		return err
	}
	defer rows.Close()

	var currentDate string
	if rows.Next() {
		if err := rows.Scan(&currentDate); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("JSON not found in the database")
	}
	// Parse the current date
	parsedDate, err := time.Parse("2006-01-02", currentDate)
	if err != nil {
		return err
	}

	// Increase the date by n days
	newDate := parsedDate.AddDate(0, 0, n)

	// Update the date in the database
	_, err = db.Exec("UPDATE main_table SET date = ? WHERE json = ?", newDate.Format(TimeFormat), jsonStr)
	return err
}

const TimeFormat string = "2006-01-02"
