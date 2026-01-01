package ui

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"snafu/data"
	"snafu/db"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func limitPathLength(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
}

func DefaultStyles() Styles {
	style := new(Styles)
	style.BorderColor = lipgloss.Color("36")
	style.InputField = lipgloss.NewStyle().BorderForeground(style.BorderColor).BorderStyle(lipgloss.NormalBorder()).Padding(1).Width(80)
	return *style
}

type Model struct {
	width         int
	height        int
	searchField   textinput.Model
	viewPort      viewport.Model
	searchResults []data.SearchResult
	style         Styles
	err           error
}

func NewModel() Model {
	styles := DefaultStyles()
	textInput := textinput.New()
	textInput.Placeholder = "Enter search term"
	textInput.Width = 30
	textInput.Focus()

	return Model{
		searchField:   textInput,
		style:         styles,
		searchResults: []data.SearchResult{},
	}
}

func (model Model) renderTable() string {
	if len(model.searchResults) == 0 {
		return ""
	}

	pathWidth := int(float64(model.width) * 0.3)
	nameWidth := int(float64(model.width) * 0.2)
	headers := []string{"Path", "Name", "Size", "Modified At", "Accessed At", "Metadata Changed"}
	var rows [][]string
	for _, entry := range model.searchResults {
		rows = append(rows, []string{
			limitPathLength(entry.Path, pathWidth),
			limitPathLength(entry.Name, nameWidth),
			strconv.Itoa(int(entry.Size)),
			time.Unix(0, entry.ModificationTime).Format("2006-01-02 15:04:05"),
			time.Unix(0, entry.AccessTime).Format("2006-01-02 15:04:05"),
			time.Unix(0, entry.MetaDataChangeTime).Format("2006-01-02 15:04:05"),
		})
	}

	table := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(model.style.BorderColor)).
		Headers(headers...).
		Rows(rows...)

	return table.Render()
}

func (model Model) Init() tea.Cmd {
	return textinput.Blink
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var viewPortCmd tea.Cmd
	model.viewPort, viewPortCmd = model.viewPort.Update(msg)
	cmds = append(cmds, viewPortCmd)

	var inputCmd tea.Cmd
	model.searchField, inputCmd = model.searchField.Update(msg)
	cmds = append(cmds, inputCmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		model.viewPort.Width = msg.Width
		model.viewPort.Height = msg.Height - 10
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return model, tea.Quit
		}
		switch msg.Type {
		case tea.KeyEnter:
			inputValue := model.searchField.Value()
			return model, searchDB(inputValue)
		}
	case searchResult:
		model.searchResults = msg.Rows
		tableString := model.renderTable()
		model.viewPort.SetContent(tableString)
		model.err = msg.Err
		return model, nil
	}
	return model, tea.Batch(cmds...)
}

func (model Model) View() string {
	if model.width == 0 {
		return "loading..."
	}

	/*var content string

	if len(model.searchResults) > 0 {
		headers := []string{"Path", "Name", "Size", "Modified At", "Accessed At", "Metadata Changed"}
		pathWidth := int(float64(model.width-10) * 0.4)
		var rows [][]string
		for _, entry := range model.searchResults {
			rows = append(rows, []string{
				limitPathLength(entry.Path, pathWidth),
				entry.Name,
				strconv.Itoa(int(entry.Size)),
				strconv.Itoa(int(entry.ModificationTime)),
				strconv.Itoa(int(entry.AccessTime)),
				strconv.Itoa(int(entry.MetaDataChangeTime)),
			})
		}

		table := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(model.style.BorderColor)).
			Headers(headers...).
			Rows(rows...)
		content = table.Render()
	}*/
	return lipgloss.Place(
		model.width,
		model.height,
		lipgloss.Center,
		lipgloss.Top,
		lipgloss.JoinVertical(
			lipgloss.Center,
			model.style.InputField.Render(model.searchField.View()),
			"\n"+model.viewPort.View(),
		),
	)

}

func searchDB(searchString string) tea.Cmd {
	return func() tea.Msg {
		homePath, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
		}
		dbPath := filepath.Join(homePath, ".snafu", "snafu.db")

		con, err := db.CreateConnection(dbPath)
		if err != nil {
			fmt.Println(err)
		}
		defer func(con *sql.DB) {
			err = db.CloseConnection(con)
			if err != nil {
				fmt.Println(err)
			}
		}(con)
		results, err := SearchIndex(con, searchString)
		if err != nil {
			fmt.Println(err)
		}
		return searchResult{Rows: results, Err: nil}
	}
}

type searchResult struct {
	Rows []data.SearchResult
	Err  error
}

func UI() {
	model := NewModel()
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	if err != nil {
		log.Fatal(err)
	}
}
