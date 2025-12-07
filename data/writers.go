package data

import (
	"database/sql"
	"fmt"
	"snafu/db"
)

func ClearExistingData(dbPath string) error {
	con, err := db.CreateConnection(dbPath)
	if err != nil {
		return err
	}
	defer func(con *sql.DB) {
		err = db.CloseConnection(con)
		if err != nil {
			fmt.Println(err)
		}
	}(con)

	query := `delete from entries;`
	_, err = con.Exec(query)
	if err != nil {
		return fmt.Errorf("could not clear existing data with query: %s\n%w", query, err)
	}

	query = `delete from ignored_entries;`
	_, err = con.Exec(query)
	if err != nil {
		return fmt.Errorf("could not clear existing data with query: %s\n%w", query, err)
	}

	return nil
}

func WriteFullEntries(dbPath string, entryCollection []*EntryCollection) error {
	con, err := db.CreateConnection(dbPath)
	if err != nil {
		return err
	}
	defer func(con *sql.DB) {
		err = db.CloseConnection(con)
		if err != nil {
			fmt.Println(err)
		}
	}(con)

	query := `insert into entries(
                    path,
					parent_directory,
					name,
					is_dir,
					size,
					modification_time,
					access_time,
					metadata_change_time,
					owner_id,
					group_id,
					extension,
					filetype,
					content_snippet,
					full_text)
					values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, entry := range entryCollection {
		_, err = con.Exec(
			query,
			entry.FullPath,
			entry.ParentDirID,
			entry.Name,
			entry.IsDir,
			entry.Size,
			entry.ModificationTime,
			entry.AccessTime,
			entry.MetaDataChangeTime,
			entry.OwnerID,
			entry.GroupID,
			entry.Extension,
			entry.FileType,
			entry.ContentSnippet,
			entry.FullTextIndex,
			entry.LineCountTotal,
			entry.LineCountWithContent)
		if err != nil {
			return fmt.Errorf("could not write entry to database: %s\n%w", query, err)
		}
	}

	return nil
}

func WriteNotRegisteredEntries(dbPath string, notRegistered []*NotAccessedPaths) error {
	con, err := db.CreateConnection(dbPath)
	if err != nil {
		return err
	}
	defer func(con *sql.DB) {
		err = db.CloseConnection(con)
		if err != nil {
			fmt.Println(err)
		}
	}(con)

	query := `insert into ignored_entries(path, error) values(?, ?)`

	for _, entry := range notRegistered {
		_, err = con.Exec(query, entry.Path, entry.Err)
		if err != nil {
			return fmt.Errorf("could not write entry to database: %s\n%w", query, err)
		}
	}

	return nil
}

func WriteScanRecord(dbPath string, theWorks *CollectedInfo) error {
	con, err := db.CreateConnection(dbPath)
	if err != nil {
		return err
	}
	defer func(con *sql.DB) {
		err = db.CloseConnection(con)
		if err != nil {
			fmt.Println(err)
		}
	}(con)

	query := `insert into full_scans(
                    scan_start,
					scan_end,
					scan_duration,
					directory_count,
				    file_count,
				    file_w_content_count,
				    ignored_entries_count,
				    indexing_completed)
					values (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = con.Exec(
		query,
		theWorks.ScanStart,
		theWorks.ScanEnd,
		theWorks.ScanDuration,
		theWorks.NumOfDirectories,
		theWorks.NumOfFiles,
		theWorks.NumOfFilesWithContent,
		theWorks.NumOfIgnoredEntries,
		theWorks.IndexingCompleted)
	if err != nil {
		return fmt.Errorf("could not write entry to database: %s\n%w", query, err)
	}
	return nil
}
