package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"stripe-seeder/internal/config"
	"stripe-seeder/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao carregar configuração: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(cfg)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
