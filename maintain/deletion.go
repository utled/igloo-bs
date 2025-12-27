package maintain

import (
	"fmt"
	"os"
	"snafu/data"
	"sync"
)

func checkDelete(entryPath string, dbPath string) error {
	if _, err := os.Stat(entryPath); err != nil {
		err = data.DeleteEntry(entryPath, dbPath)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func traverseIndexedEntries(deletionJobs chan<- string, inodeMappedEntries map[uint64]data.InodeHeader, wg *sync.WaitGroup) error {
	defer wg.Done()
	defer close(deletionJobs)

	for _, values := range inodeMappedEntries {
		deletionJobs <- values.Path
	}
	return nil
}
