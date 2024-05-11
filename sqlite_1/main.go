package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var dbConn *sql.DB

func main() {
	var err error

	// Connection
	dbConn, err = sql.Open("sqlite3", ".db")
	if err != nil {
		slog.Error("Error connecting to the database", "Err", err)
		return
	}
	defer dbConn.Close()

	// Table creation
	_, err = dbConn.Exec("CREATE TABLE config_table (name TEXT NOT NULL UNIQUE, value TEXT);")
	if err != nil {
		// Probably table already exists
		err = nil // Something to get rid of empty branch warning :)
		// slog.Warn("Error creating new table", "Err", err)
	}

	// Get count of elements in the table
	var count int
	var rows *sql.Rows
	rows, err = dbConn.Query("SELECT COUNT(*) FROM config_table;")
	if err != nil {
		slog.Error("Error querying count of elements in the table", "Err", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			break
		}
		fmt.Printf("Count of records in the table: %d\n", count)
	}

	// Add new record
	_, err = dbConn.Exec(fmt.Sprintf("INSERT INTO config_table (name, value) VALUES ('%s', '%s');",
		fmt.Sprintf("Variable%d", count+1),
		fmt.Sprintf("Value%d", count+1)))
	if err != nil {
		slog.Error("Error when inserting new record", "Err", err)
	}

	// Read all records
	rows, err = dbConn.Query("SELECT * FROM config_table;")
	if err != nil {
		slog.Error("Error querying elements of elements in the table", "Err", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, value string
		err = rows.Scan(&name, &value)
		if err != nil {
			break
		}
		fmt.Printf("Added new record, Name: %s, Value: %s\n", name, value)
	}

	// Query value or create new one
	var val string
	var randomName = uuid.NewString()
	val, err = GetValueOrCreteNew(randomName, uuid.NewString())
	if err != nil {
		slog.Error("Error when requesting record value", "Err", err)
		return
	}
	fmt.Printf("ValueOrNew result, Name: %s, Value: %s\n", randomName, val)
}

func GetValueOrCreteNew(name string, defaultValue string) (string, error) {
	if dbConn == nil {
		return "", errors.New("database connection not establised")
	}

	var err error
	var val string
	err = dbConn.QueryRow(fmt.Sprintf("SELECT value FROM config_table WHERE name='%s' LIMIT 1;", name)).Scan(&val)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// Create new record
			_, err = dbConn.Exec(fmt.Sprintf("INSERT INTO config_table (name, value) VALUES ('%s', '%s');", name, defaultValue))
			if err != nil {
				return "", err
			}
			val = defaultValue
		} else {
			return "", err
		}
	}

	return val, nil
}
