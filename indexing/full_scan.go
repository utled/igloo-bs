package indexing

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"snafu/data"
	"sync"
	"syscall"
	"time"
)

const (
	directoryWorkers       = 20
	fileWorkers            = 80
	directoryJobBufferSize = 100
	fileJobBufferSize      = 500
)

func readDir(path string, theWorks *data.CollectedInfo, isRoot bool) {
	entry := data.EntryCollection{}

	dirStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	entry.FullPath = path
	if !isRoot {
		entry.ParentDirID = filepath.Dir(path)
	}
	entry.Name = filepath.Base(path)
	entry.IsDir = true
	entry.Size = dirStat.Size()

	statT := dirStat.Sys().(*syscall.Stat_t)
	entry.ModificationTime = statT.Mtim.Sec + statT.Mtim.Nsec
	entry.AccessTime = statT.Atim.Sec + statT.Atim.Nsec
	entry.MetaDataChangeTime = statT.Ctim.Sec + statT.Ctim.Nsec

	entry.OwnerID = statT.Uid
	entry.GroupID = statT.Gid
	entry.Extension = filepath.Ext(entry.Name)
	entry.FileType = filepath.Ext(entry.Name)

	theWorks.Mu.Lock()
	theWorks.NumOfDirectories += 1
	theWorks.EntryDetails = append(theWorks.EntryDetails, &entry)
	theWorks.Mu.Unlock()
}

func readFile(filename string, theWorks *data.CollectedInfo) {
	entry := data.EntryCollection{}

	contentsRead := false

	contentFiles := []string{".txt", ".md", ".go", ".py"}
	if slices.Contains(contentFiles, filepath.Ext(filename)) {
		contents, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		contentsRead = true
		lineCountTotal := bytes.Count(contents, []byte("\n"))
		blankLines := bytes.Count(contents, []byte("\n\n"))
		lineCountWithContent := lineCountTotal - blankLines

		contents = bytes.ReplaceAll(contents, []byte("\n"), []byte(" "))
		contents = bytes.ReplaceAll(contents, []byte("\r"), []byte(" "))
		contents = bytes.ReplaceAll(contents, []byte("\t"), []byte(" "))

		regExCleanup := regexp.MustCompile(`[\p{C}\p{Zl}\p{Zp}]`)
		contents = regExCleanup.ReplaceAll(contents, []byte(" "))
		contents = regexp.MustCompile(`\s+`).ReplaceAll(contents, []byte(" "))
		if len(contents) < 500 {
			entry.ContentSnippet = contents
		} else {
			entry.ContentSnippet = contents[:500]
		}
		entry.FullTextIndex = contents
		entry.LineCountTotal = lineCountTotal
		entry.LineCountWithContent = lineCountWithContent
	}

	fileStat, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	entry.FullPath = filename
	entry.ParentDirID = filepath.Dir(filename)
	entry.Name = filepath.Base(filename)
	entry.IsDir = false
	entry.Size = fileStat.Size()

	statT := fileStat.Sys().(*syscall.Stat_t)
	entry.ModificationTime = statT.Mtim.Sec + statT.Mtim.Nsec
	entry.AccessTime = statT.Atim.Sec + statT.Atim.Nsec
	entry.MetaDataChangeTime = statT.Ctim.Sec + statT.Ctim.Nsec

	entry.OwnerID = statT.Uid
	entry.GroupID = statT.Gid

	theWorks.Mu.Lock()
	theWorks.NumOfFiles += 1
	if contentsRead {
		theWorks.NumOfFilesWithContent += 1
	}
	theWorks.EntryDetails = append(theWorks.EntryDetails, &entry)
	theWorks.Mu.Unlock()
}

func traverseDirectory(root string, dirJobs chan<- data.ReadJob, fileJobs chan<- data.ReadJob, wg *sync.WaitGroup, theWorks *data.CollectedInfo) {
	defer wg.Done()

	defer close(dirJobs)
	defer close(fileJobs)

	excludedEntries := []string{
		".cache",
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			failedPath := data.NotAccessedPaths{Path: path, Err: err.Error()}
			theWorks.Mu.Lock()
			theWorks.NumOfIgnoredEntries += 1
			theWorks.NotRegistered = append(theWorks.NotRegistered, &failedPath)
			theWorks.Mu.Unlock()
			return nil
		}

		if path == root {
			return nil
		}

		_, err = os.Stat(path)
		if err != nil {
			return nil
		}

		if d.IsDir() && slices.Contains(excludedEntries, filepath.Base(path)) {
			return filepath.SkipDir
		}

		job := data.ReadJob{Path: path}
		if d.IsDir() {
			dirJobs <- job
		} else {
			fileJobs <- job
		}

		return nil
	})

	if err != nil {
		log.Printf("Fatal error during directory traversal: %v", err)
	}
}

func dirWorker(readJobs <-chan data.ReadJob, wg *sync.WaitGroup, theWorks *data.CollectedInfo) {
	defer wg.Done()

	for t := range readJobs {
		readDir(t.Path, theWorks, false)
	}
}

func fileWorker(readJobs <-chan data.ReadJob, wg *sync.WaitGroup, theWorks *data.CollectedInfo) {
	defer wg.Done()
	for t := range readJobs {
		readFile(t.Path, theWorks)
	}
}

func Main() {
	start := time.Now()
	theWorks := data.CollectedInfo{}

	fileReadJobs := make(chan data.ReadJob, fileJobBufferSize)
	dirReadJobs := make(chan data.ReadJob, directoryJobBufferSize)

	var wg sync.WaitGroup
	totalWorkers := 1 + directoryWorkers + fileWorkers
	wg.Add(totalWorkers)

	path := "/home/utled"
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
