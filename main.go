package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"db-tui/internal/app"
)

func main() {
	m := app.New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
