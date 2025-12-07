package db

import (
	"database/sql"
	"fmt"
)

type DefaultConfig struct{}

func InitializeDB() error {
	db, err := CreateConnection()
	if err != nil {
		return err
	}
	defer CloseConnection(db)

	err = createTables(db)
	if err != nil {
		return err
	}

	err = writeDefaultConfig(db)
	if err != nil {
		return err
	}

	return nil
}

func createTables(db *sql.DB) error {
	statements := []string{
		`create table if not exists config ();`,
		`create table if not exists full_scans ();`,
		`create table if not exists entries ();`,
		`create table if not exists changes ();`,
	}

	for _, statement := range statements {
		_, err := db.Exec(statement)
		if err != nil {
			return fmt.Errorf("could not create table %s: \n%w", statement, err)
		}
	}

	return nil
}

func writeDefaultConfig(db *sql.DB) error {
	query := ""
	queryResponse := db.QueryRow(query)

	config := &DefaultConfig{}
	err := queryResponse.Scan(&config)
	if err != nil {
		if err == sql.ErrNoRows {
			// set default values for config.values
			_, err = db.Exec(query)
			if err != nil {
				return fmt.Errorf("failed to write default config: \n%w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to read default config: \n%w", err)
		}
	}

	return nil
}
