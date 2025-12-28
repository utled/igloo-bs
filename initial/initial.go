package initial

import (
	"fmt"
	"log"
	"os"
	"snafu/data"
	"sync"
	"time"
)

const (
	directoryWorkers       = 20
	fileWorkers            = 80
	directoryJobBufferSize = 100
	fileJobBufferSize      = 500
)

func StartInitialScan() {
	start := time.Now()
	theWorks := data.CollectedInfo{}

	fileReadJobs := make(chan string, fileJobBufferSize)
	dirReadJobs := make(chan string, directoryJobBufferSize)

	var wg sync.WaitGroup
	totalWorkers := 1 + directoryWorkers + fileWorkers
	wg.Add(totalWorkers)

	path := "/home/utled/GolandProjects"
	stat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if !stat.IsDir() {
		log.Fatal("Starting path must be a directory")
	}
	readDir(path, &theWorks, true)

	for i := 0; i < directoryWorkers; i += 1 {
		go dirWorker(dirReadJobs, &wg, &theWorks)
	}

	for i := 0; i < fileWorkers; i += 1 {
		go fileWorker(fileReadJobs, &wg, &theWorks)
	}

	go traverseDirectory(path, dirReadJobs, fileReadJobs, &wg, &theWorks)

	wg.Wait()
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Printf("Full scan took %s\n", elapsed)

	theWorks.Mu.Lock()
	theWorks.ScanStart = start
	theWorks.ScanEnd = end
	theWorks.ScanDuration = elapsed
	theWorks.Mu.Unlock()

	writeStart := time.Now()
	err = updateFullIndex(&theWorks)
	if err != nil {
		fmt.Println(err)
	}
	writeElapsed := time.Since(writeStart)
	fmt.Printf("Full dbWrite took %s\n", writeElapsed)
}
