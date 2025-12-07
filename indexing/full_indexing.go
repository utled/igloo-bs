package indexing

import (
	"os"
	"path/filepath"
	"snafu/data"
)

func updateFullIndex(theWorks *data.CollectedInfo) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dbPath := filepath.Join(homePath, ".snafu", "snafu.db")

	err = data.ClearExistingData(dbPath)
	if err != nil {
		return err
	}

	err = data.WriteFullEntries(dbPath, theWorks.EntryDetails)
	if err != nil {
		return err
	}

	err = data.WriteNotRegisteredEntries(dbPath, theWorks.NotRegistered)
	if err != nil {
		return err
	}

	theWorks.IndexingCompleted = true

	err = data.WriteScanRecord(dbPath, theWorks)
	if err != nil {
		return err
	}

	return nil
}
