package main

import (
	"fmt"
	"os"
	"path/filepath"

	"pos/internal/database"
	"pos/internal/store"
	"pos/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".pos", "pos.db")

	db, err := database.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error abriendo base de datos: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	s := store.New(db)
	app := tui.NewApp(s)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
