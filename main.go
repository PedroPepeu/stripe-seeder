package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"stripe-seeder/internal/config"
	"stripe-seeder/internal/logger"
	"stripe-seeder/internal/tui"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logging to ~/.stripe-seeder-debug.log")
	flag.Parse()

	if *debug {
		if err := logger.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Aviso: debug log não iniciado: %v\n", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao carregar configuração: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(cfg, *debug)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
