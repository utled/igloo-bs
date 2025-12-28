package maintain

import (
	"fmt"
	"os"
	"path/filepath"
	"snafu/data"
	"syscall"
)

func scanUpdatedDir(readJobs chan<- data.SyncJob, dirPath string, inodeMappedEntries map[uint64]data.InodeHeader) error {
	fileSysEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to list entries in directory: %s\n%w", dirPath, err)
	}

	for _, entry := range fileSysEntries {
		filePath := filepath.Join(dirPath, entry.Name())
		entryStat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		entryStatT := entryStat.Sys().(*syscall.Stat_t)
		entryMtim := entryStatT.Mtim.Sec + entryStatT.Mtim.Nsec

		if inode, ok := inodeMappedEntries[entryStatT.Ino]; !ok {
			syncJob := data.SyncJob{Path: filePath, IsIndexed: false, IsContentChange: !entry.IsDir()}
			readJobs <- syncJob
		} else {
			if entryMtim != inode.ModificationTime {
				syncJob := data.SyncJob{Path: filePath, IsIndexed: true, IsContentChange: !entry.IsDir()}
				readJobs <- syncJob
			} else {
				syncJob := data.SyncJob{Path: filePath, IsIndexed: true, IsContentChange: false}
				readJobs <- syncJob
			}
		}
	}

	return nil
}
