package data

import (
	"database/sql"
	"fmt"
)

func GetInodeMappedEntries(con *sql.DB) (inodeMappedEntries map[uint64]InodeHeader, err error) {
	inodeMappedEntries = make(map[uint64]InodeHeader)
	if err != nil {
		return inodeMappedEntries, err
	}
	var query string
	var response *sql.Rows
	query = `select inode, path, modification_time, metadata_change_time 
				from entries
				order by inode;`
	response, err = con.Query(query)

	for response.Next() {
		var inode uint64
		var details InodeHeader
		err = response.Scan(
			&inode,
			&details.Path,
			&details.ModificationTime,
			&details.MetaDataChangeTime,
		)
		if err != nil {
			return inodeMappedEntries, fmt.Errorf("failed to serialize entry details to map: %v", err)
		}
		inodeMappedEntries[inode] = details
	}
	if err = response.Err(); err != nil {
		return inodeMappedEntries, fmt.Errorf("failed to iterate through db response: %v", err)
	}

	return inodeMappedEntries, nil
}
