package maintain

import (
	"database/sql"
	"log"
	"snafu/data"
	"sync"
)

func scanWorker(scanJobs <-chan data.InodeHeader, readJobs chan<- data.SyncJob, inodeMappedEntries map[uint64]data.InodeHeader, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range scanJobs {
		err := scanUpdatedDir(readJobs, job.Path, inodeMappedEntries)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func readWorker(readJobs <-chan data.SyncJob, con *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range readJobs {
		readEntry(job, con)
	}
}
func newDirWorker(newDirJobs <-chan string, readJobs chan<- data.SyncJob, con *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for path := range newDirJobs {
		err := traverseNewDir(readJobs, path, con)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func deletionWorker(delJobs <-chan string, con *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for path := range delJobs {
		err := checkDelete(path, con)
		if err != nil {
			log.Fatal(err)
		}
	}
}
