package maintain

import (
	"fmt"
	"os"
	"path/filepath"
	"snafu/data"
	"sync"
)

const (
	deletionJobBufferSize = 100
	scanJobBufferSize     = 100
	readJobBufferSize     = 500
	newDirJobBufferSize   = 100
	deletionWorkers       = 20
	entryScanners         = 20
	entryReaders          = 80
	newDirWorkers         = 20
)

func orchestrateScan(startPath string) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dbPath := filepath.Join(homePath, ".snafu", "snafu.db")
	inodeMappedEntries, err := data.GetInodeMappedEntries(dbPath)
	if err != nil {
		fmt.Println(err)
	}

	deletionJobs := make(chan string, deletionJobBufferSize)
	scanJobs := make(chan data.InodeHeader, scanJobBufferSize)
	newDirJobs := make(chan string, newDirJobBufferSize)
	readJobs := make(chan data.SyncJob, readJobBufferSize)

	var deletionProdWG sync.WaitGroup
	var deletionWG sync.WaitGroup
	var producerWG sync.WaitGroup
	var scannerWG sync.WaitGroup
	var readerWG sync.WaitGroup

	deletionWG.Add(deletionWorkers)
	for i := 0; i < deletionWorkers; i++ {
		go deletionWorker(deletionJobs, dbPath, &deletionWG)
	}

	deletionProdWG.Add(1)
	traverseIndexedEntries(deletionJobs, inodeMappedEntries, &deletionProdWG)

	scannerWG.Add(entryScanners)
	for i := 0; i < entryScanners; i += 1 {
		go scanWorker(scanJobs, readJobs, inodeMappedEntries, &scannerWG)
	}

	scannerWG.Add(newDirWorkers)
	for i := 0; i < newDirWorkers; i += 1 {
		go newDirWorker(newDirJobs, readJobs, dbPath, &scannerWG)
	}

	readerWG.Add(entryReaders)
	for i := 0; i < entryReaders; i += 1 {
		go readWorker(readJobs, dbPath, &readerWG)
	}

	producerWG.Add(1)
	go traverseDirectories(startPath, scanJobs, newDirJobs, readJobs, &producerWG, inodeMappedEntries)

	producerWG.Wait()
	close(scanJobs)
	close(newDirJobs)

	scannerWG.Wait()
	close(readJobs)

	readerWG.Wait()
	deletionProdWG.Wait()
	deletionWG.Wait()

	return nil
}
