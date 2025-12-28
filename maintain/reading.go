package maintain

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"snafu/data"
	"snafu/utils"
	"syscall"
)

func readEntry(syncJob data.SyncJob, dbPath string) {
	entryStat, err := os.Stat(syncJob.Path)
	if err != nil {
		fmt.Println(err)
	}
	entry := data.EntryCollection{}

	entry.FullPath = syncJob.Path
	entry.ParentDirID = filepath.Dir(syncJob.Path)
	entry.Name = filepath.Base(syncJob.Path)
	entry.IsDir = entryStat.IsDir()
	entry.Size = entryStat.Size()

	statT := entryStat.Sys().(*syscall.Stat_t)
	entry.Inode = statT.Ino
	entry.ModificationTime = statT.Mtim.Sec + statT.Mtim.Nsec
	entry.AccessTime = statT.Atim.Sec + statT.Atim.Nsec
	entry.MetaDataChangeTime = statT.Ctim.Sec + statT.Ctim.Nsec

	entry.OwnerID = statT.Uid
	entry.GroupID = statT.Gid

	if !entryStat.IsDir() {
		if slices.Contains(utils.ContentFiles, filepath.Ext(syncJob.Path)) && syncJob.IsContentChange {
			contents, err := os.ReadFile(syncJob.Path)
			if err != nil {
				log.Fatal(err)
			}
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
	}

	entryCollection := make([]*data.EntryCollection, 1)
	entryCollection[0] = &entry
	if !syncJob.IsIndexed {
		err := data.WriteFullEntries(dbPath, entryCollection)
		if err != nil {
			fmt.Println("error writing: ", entry.FullPath, err)
		}
		return
	}
	if syncJob.IsContentChange {
		err := data.UpdateEntriesWithContent(dbPath, entryCollection)
		if err != nil {
			fmt.Println("error updating: ", entry.FullPath, err)
		}
		return
	}
	err = data.UpdateEntriesWithoutContent(dbPath, entryCollection)
	if err != nil {
		fmt.Println("error updating: ", entry.FullPath, err)
	}
}
