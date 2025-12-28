package initial

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"snafu/data"
	"snafu/db"
)

func updateFullIndex(theWorks *data.CollectedInfo) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dbPath := filepath.Join(homePath, ".snafu", "snafu.db")
	con, err := db.CreateConnection(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(con *sql.DB) {
		err = db.CloseConnection(con)
		if err != nil {
			fmt.Println(err)
		}
	}(con)

	err = data.ClearExistingData(con)
	if err != nil {
		return err
	}

	err = data.WriteFullEntries(con, theWorks.EntryDetails)
	if err != nil {
		return err
	}

	err = data.WriteNotRegisteredEntries(con, theWorks.NotRegistered)
	if err != nil {
		return err
	}

	theWorks.IndexingCompleted = true

	err = data.WriteScanRecord(con, theWorks)
	if err != nil {
		return err
	}

	return nil
}
