package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type dbHandler struct {
	db *sql.DB
}

// initDB initializes Database using Sqlite3
func (dh *dbHandler) initDB() error {
	// Open database connection.
	var err error
	dh.db, err = sql.Open("sqlite3", "./events.db")
	if err != nil {
		return err
	}

	// Create events table if it does not exist.
	createStmt := `
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY,
			datetime TEXT NOT NULL,
			hostname TEXT NOT NULL,
			pid INTEGER NOT NULL,
			ppid INTEGER NOT NULL,
			ppName TEXT NOT NULL,
			uid INTEGER NOT NULL,
			username TEXT NOT NULL,
			command TEXT NOT NULL,
			container TEXT NOT NULL,
			isContainer INTEGER NOT NULL,
			podName TEXT NOT NULL
		);
	`
	_, err = dh.db.Exec(createStmt)
	if err != nil {
		return err
	}

	return nil
}

// insertEvent inserts event into the sqlite using insert operation.
func (dh *dbHandler) insertEvent(event eventInfo) error {
	var err error

	// Insert event into the database.
	insertStmt := `
		INSERT INTO events (datetime, hostname, pid, ppid, ppName, uid, username, command, container, isContainer, podName)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = dh.db.Exec(insertStmt, time.Now().Format("2006-01-02 15:04:05"), event.Hostname, event.Pid, event.Ppid, event.PpName,
		event.Uid, event.Username, event.Command, event.Container, event.IsContainer, event.PodName)
	if err != nil {
		return err
	}

	return nil
}
