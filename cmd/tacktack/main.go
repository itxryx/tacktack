package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/itxryx/tacktack/internal/model"
)

func main() {
	dbPath, err := db.ResolveDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: DBパスの解決に失敗しました: %v\n", err)
		os.Exit(1)
	}

	gormDB, err := db.InitDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: DB接続に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := db.Migrate(gormDB); err != nil {
		fmt.Fprintf(os.Stderr, "Error: マイグレーションに失敗しました: %v\n", err)
		os.Exit(1)
	}

	m := model.New(gormDB)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
