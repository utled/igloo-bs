package data

import (
	"sync"
	"time"
)

type CollectedInfo struct {
	ScanStart             time.Time
	ScanEnd               time.Time
	ScanDuration          time.Duration
	IndexingCompleted     bool
	NumOfFiles            int
	NumOfDirectories      int
	NumOfFilesWithContent int
	NumOfIgnoredEntries   int
	EntryDetails          []*EntryCollection
	NotRegistered         []*NotAccessedPaths
	Mu                    sync.Mutex
}

type EntryCollection struct {
	FullPath    string // primary key
	ParentDirID string // foreign key
	Name        string
	IsDir       bool
	Size        int64
	//creationTime       int64 // Btim (not included syscall.Stat_t)
	ModificationTime     int64  // os.fileStat.modTime or os.fileStat.sys.Mtim.Sec + Mtim.Nsec
	AccessTime           int64  // os.fileStat.sys.Atim.Sec + Atim.Nsec
	MetaDataChangeTime   int64  // os.fileStat.sys.Ctim.Sec + Ctim.Nsec
	OwnerID              uint32 // os.fileStat.sys.Uid
	GroupID              uint32 // os.fileStat.sys.Gid
	Extension            string
	FileType             string // MIME type. Determined by file extension and/or internal magic bytes
	ContentSnippet       []byte // short extract of the files content. [:500] to start with
	FullTextIndex        []byte // the complete textual content of a document, stored in separate Full-Text Search index
	LineCountTotal       int
	LineCountWithContent int
	//tags               []string // user defined tags or keywords from internal metadata
}

type NotAccessedPaths struct {
	Path string
	Err  string
}

type ReadJob struct {
	Path string
}
