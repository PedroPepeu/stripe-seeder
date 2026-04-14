package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"stripe-seeder/internal/config"
	stripeClient "stripe-seeder/internal/stripe"
)

// --- Screen states -----------------------------------------------------------

type screen int

const (
	screenMain screen = iota
	screenSetKey
	screenSeedProducts
	screenSeedProductsPrices
	screenSeedCustomers
	screenResults
	screenLoading
)

// --- Menu items --------------------------------------------------------------

type menuItem struct {
	title string
	desc  string
}

var mainMenuItems = []menuItem{
	{title: "🔗  Ver Projeto Conectado", desc: "Mostra a chave API e ambiente atual"},
	{title: "🔑  Alterar Chave API", desc: "Configurar ou trocar a Stripe API Key"},
	{title: "📦  Seed: Produtos", desc: "Criar produtos com nomes aleatórios"},
	{title: "💰  Seed: Produtos + Preços", desc: "Criar produtos com preços aleatórios"},
	{title: "👤  Seed: Clientes", desc: "Criar clientes com dados aleatórios"},
	{title: "🚪  Sair", desc: "Encerrar o programa"},
}

// --- Messages ----------------------------------------------------------------

type seedDoneMsg struct {
	result stripeClient.SeedResult
	label  string
}

type validateKeyMsg struct {
	env string
	err error
}

// --- Model -------------------------------------------------------------------

type Model struct {
	cfg     *config.Config
	screen  screen
	cursor  int
	env     string // "TEST 🟡" or "LIVE 🔴"

	// text inputs for forms
	inputs     []textinput.Model
	inputFocus int

	// loading
	spinner spinner.Model
	loading bool

	// results
	resultTitle string
	resultLines []string

	// status message
	statusMsg string

	width  int
	height int
}

func NewModel(cfg *config.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	m := Model{
		cfg:     cfg,
		screen:  screenMain,
		spinner: s,
	}

	// try to detect env from saved key
	if cfg.APIKey != "" {
		if strings.HasPrefix(cfg.APIKey, "sk_test") {
			m.env = "TEST 🟡"
		} else if strings.HasPrefix(cfg.APIKey, "sk_live") {
			m.env = "LIVE 🔴"
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// --- Update ------------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case seedDoneMsg:
		m.loading = false
		m.screen = screenResults
		m.resultTitle = msg.label
		m.resultLines = msg.result.Details
		m.statusMsg = fmt.Sprintf("✓ %d criados  ✗ %d erros", msg.result.Created, msg.result.Errors)
		return m, nil

	case validateKeyMsg:
		m.loading = false
		if msg.err != nil {
			m.statusMsg = errorStyle.Render("✗ " + msg.err.Error())
			m.screen = screenMain
		} else {
			m.env = msg.env
			m.statusMsg = successStyle.Render(fmt.Sprintf("✓ Conectado! Ambiente: %s", msg.env))
			m.screen = screenMain
		}
		return m, nil
	}

	switch m.screen {
	case screenMain:
		return m.updateMain(msg)
	case screenSetKey:
		return m.updateSetKey(msg)
	case screenSeedProducts:
		return m.updateSeedProducts(msg)
	case screenSeedProductsPrices:
		return m.updateSeedProductsPrices(msg)
	case screenSeedCustomers:
		return m.updateSeedCustomers(msg)
	case screenResults:
		return m.updateResults(msg)
	case screenLoading:
		return m, nil
	}
	return m, nil
}

func (m Model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "q" || msg.String() == "ctrl+c":
			return m, tea.Quit
		case msg.String() == "up" || msg.String() == "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case msg.String() == "down" || msg.String() == "j":
			if m.cursor < len(mainMenuItems)-1 {
				m.cursor++
			}
		case msg.String() == "enter":
			return m.handleMenuSelect()
		}
	}
	return m, nil
}

func (m Model) handleMenuSelect() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0: // ver projeto
		if m.cfg.APIKey == "" {
			m.statusMsg = warningStyle.Render("⚠ Nenhuma chave configurada. Use 'Alterar Chave API'.")
		} else {
			masked := maskKey(m.cfg.APIKey)
			m.statusMsg = fmt.Sprintf("Chave: %s  |  Ambiente: %s", masked, m.env)
		}
		return m, nil

	case 1: // set key
		m.screen = screenSetKey
		m.inputs = make([]textinput.Model, 2)

		ti := textinput.New()
		ti.Placeholder = "sk_test_..."
		ti.CharLimit = 200
		ti.Width = 50
		ti.Focus()
		if m.cfg.APIKey != "" {
			ti.SetValue(m.cfg.APIKey)
		}
		m.inputs[0] = ti

		ti2 := textinput.New()
		ti2.Placeholder = "Meu Projeto"
		ti2.CharLimit = 60
		ti2.Width = 50
		if m.cfg.ProjectName != "" {
			ti2.SetValue(m.cfg.ProjectName)
		}
		m.inputs[1] = ti2

		m.inputFocus = 0
		m.statusMsg = ""
		return m, textinput.Blink

	case 2: // seed products
		if m.cfg.APIKey == "" {
			m.statusMsg = warningStyle.Render("⚠ Configure a chave API primeiro!")
			return m, nil
		}
		m.screen = screenSeedProducts
		m.inputs = make([]textinput.Model, 1)

		ti := textinput.New()
		ti.Placeholder = "10"
		ti.CharLimit = 5
		ti.Width = 20
		ti.Focus()
		m.inputs[0] = ti
		m.inputFocus = 0
		m.statusMsg = ""
		return m, textinput.Blink

	case 3: // seed products + prices
		if m.cfg.APIKey == "" {
			m.statusMsg = warningStyle.Render("⚠ Configure a chave API primeiro!")
			return m, nil
		}
		m.screen = screenSeedProductsPrices
		m.inputs = make([]textinput.Model, 4)

		fields := []struct {
			placeholder string
			width       int
		}{
			{"10", 20},
			{"500", 20},
			{"50000", 20},
			{"brl", 20},
		}
		for i, f := range fields {
			ti := textinput.New()
			ti.Placeholder = f.placeholder
			ti.CharLimit = 20
			ti.Width = f.width
			m.inputs[i] = ti
		}
		m.inputs[0].Focus()
		m.inputFocus = 0
		m.statusMsg = ""
		return m, textinput.Blink

	case 4: // seed customers
		if m.cfg.APIKey == "" {
			m.statusMsg = warningStyle.Render("⚠ Configure a chave API primeiro!")
			return m, nil
		}
		m.screen = screenSeedCustomers
		m.inputs = make([]textinput.Model, 1)

		ti := textinput.New()
		ti.Placeholder = "10"
		ti.CharLimit = 5
		ti.Width = 20
		ti.Focus()
		m.inputs[0] = ti
		m.inputFocus = 0
		m.statusMsg = ""
		return m, textinput.Blink

	case 5: // quit
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateSetKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "tab", "shift+tab":
			m.inputFocus = (m.inputFocus + 1) % len(m.inputs)
			for i := range m.inputs {
				if i == m.inputFocus {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, textinput.Blink
		case "enter":
			apiKey := strings.TrimSpace(m.inputs[0].Value())
			projectName := strings.TrimSpace(m.inputs[1].Value())
			if apiKey == "" {
				m.statusMsg = errorStyle.Render("✗ A chave API não pode estar vazia.")
				return m, nil
			}
			m.cfg.APIKey = apiKey
			if projectName != "" {
				m.cfg.ProjectName = projectName
			}
			_ = config.Save(m.cfg)

			m.screen = screenLoading
			m.loading = true
			m.statusMsg = "Validando chave..."

			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				env, err := stripeClient.ValidateKey(apiKey)
				return validateKeyMsg{env: env, err: err}
			})
		}
	}

	var cmd tea.Cmd
	m.inputs[m.inputFocus], cmd = m.inputs[m.inputFocus].Update(msg)
	return m, cmd
}

func (m Model) updateSeedProducts(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "enter":
			n := parseIntOr(m.inputs[0].Value(), 10)
			m.screen = screenLoading
			m.loading = true
			m.statusMsg = fmt.Sprintf("Criando %d produtos...", n)
			apiKey := m.cfg.APIKey
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				res := stripeClient.SeedProducts(apiKey, n)
				return seedDoneMsg{result: res, label: "Seed: Produtos"}
			})
		}
	}
	var cmd tea.Cmd
	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func (m Model) updateSeedProductsPrices(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "tab", "shift+tab":
			m.inputFocus = (m.inputFocus + 1) % len(m.inputs)
			for i := range m.inputs {
				if i == m.inputFocus {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, textinput.Blink
		case "enter":
			n := parseIntOr(m.inputs[0].Value(), 10)
			minC := int64(parseIntOr(m.inputs[1].Value(), 500))
			maxC := int64(parseIntOr(m.inputs[2].Value(), 50000))
			cur := strings.ToLower(strings.TrimSpace(m.inputs[3].Value()))
			if cur == "" {
				cur = "brl"
			}
			if minC > maxC {
				minC, maxC = maxC, minC
			}

			m.screen = screenLoading
			m.loading = true
			m.statusMsg = fmt.Sprintf("Criando %d produtos + preços...", n)
			apiKey := m.cfg.APIKey
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				res := stripeClient.SeedProductsWithPrices(apiKey, n, minC, maxC, cur)
				return seedDoneMsg{result: res, label: "Seed: Produtos + Preços"}
			})
		}
	}
	var cmd tea.Cmd
	m.inputs[m.inputFocus], cmd = m.inputs[m.inputFocus].Update(msg)
	return m, cmd
}

func (m Model) updateSeedCustomers(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "enter":
			n := parseIntOr(m.inputs[0].Value(), 10)
			m.screen = screenLoading
			m.loading = true
			m.statusMsg = fmt.Sprintf("Criando %d clientes...", n)
			apiKey := m.cfg.APIKey
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				res := stripeClient.SeedCustomers(apiKey, n)
				return seedDoneMsg{result: res, label: "Seed: Clientes"}
			})
		}
	}
	var cmd tea.Cmd
	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func (m Model) updateResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter", "q":
			m.screen = screenMain
			return m, nil
		}
	}
	return m, nil
}

// --- Helpers -----------------------------------------------------------------

func maskKey(key string) string {
	if len(key) <= 12 {
		return "****"
	}
	return key[:7] + "..." + key[len(key)-4:]
}

func parseIntOr(s string, fallback int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
