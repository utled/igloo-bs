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
	tableStatements := []string{
		`create table if not exists config (
    		type text unique not null primary key,
    		priority integer not null default 0
		);`,
		`create table if not exists full_scans (
    		scan_id int auto_increment primary key,
         	scan_start text,
         	scan_end text,
         	scan_duration text,
         	directory_count int,
         	file_count int,
         	file_w_content_count int,
         	ignored_entries_count int
         );`,
		`create table if not exists entries (
    		path text not null,
    		parent_directory text,
    		name text,
    		is_dir boolean,
    		size int,
    		modification_time int,
    		access_time int,
    		metadata_change_time int,
    		owner_id int,
    		group_id int,
    		extension text,
    		filetype text,
    		content_snippet text,
    		full_text text
		);`,
		`create table if not exists ignored_entries (
    		path text,
    		error text
		);`,
		`create table if not exists changes (
    		path text,
    		field text,
    		before text,
        	after text
		);`,
	}

	for _, statement := range tableStatements {
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
