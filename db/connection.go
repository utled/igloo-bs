package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func CreateConnection() (*sql.DB, error) {
	dbPath := "/home/utled/.snafu.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	return db, nil
}

func CloseConnection(db *sql.DB) error {
	err := db.Close()
	if err != nil {
		return fmt.Errorf("faIled to close db connection: %v", err)
	}

	return nil
}
