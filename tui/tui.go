package tui

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"igloo/db"

	tea "github.com/charmbracelet/bubbletea"
)

func UI() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	indexDBPath := filepath.Join(homePath, ".icu", "icu.db")

	con, err := db.CreateConnection(indexDBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(con *sql.DB) {
		err = db.CloseConnection(con)
		if err != nil {
			log.Fatal(err)
		}
	}(con)

	model := NewModel(con)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	if err != nil {
		log.Fatal(err)
	}
}
