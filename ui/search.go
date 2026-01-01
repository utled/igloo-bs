package ui

import (
	"database/sql"
	"fmt"
	"snafu/data"
)

func SearchIndex(con *sql.DB, searchString string) (searchResults []data.SearchResult, err error) {
	var query string
	var response *sql.Rows

	query = `select path, name, size, modification_time, access_time, metadata_change_time 
				from entries 
				where name like ?;`
	response, err = con.Query(query, searchString+"%")
	for response.Next() {
		var entry data.SearchResult
		err = response.Scan(
			&entry.Path,
			&entry.Name,
			&entry.Size,
			&entry.ModificationTime,
			&entry.AccessTime,
			&entry.MetaDataChangeTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize search results: %v", err)
		}
		searchResults = append(searchResults, entry)
	}
	if err = response.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate through search results: %v", err)
	}

	return searchResults, nil
}
