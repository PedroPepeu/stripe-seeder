package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PedroPepeu/stripe-seeder/internal/config"
	"github.com/PedroPepeu/stripe-seeder/internal/logger"
	stripeClient "github.com/PedroPepeu/stripe-seeder/internal/stripe"
)

// --- Screen states -----------------------------------------------------------

type screen int

const (
	screenMain screen = iota
	screenLogin
	screenSeedProducts
	screenSeedProductsPrices
	screenSeedCustomers
	screenSeedPaymentIntents
	screenResults
	screenLoading
	screenDebugLog
)

// --- Menu items --------------------------------------------------------------

type menuItem struct {
	title string
	desc  string
}

// --- Messages ----------------------------------------------------------------

type seedDoneMsg struct {
	result stripeClient.SeedResult
	label  string
}

type checkLoginMsg struct {
	info string
	err  error
}

type loginDoneMsg struct{ err error }

type loadLogMsg struct{ content string }

// --- Model -------------------------------------------------------------------

type Model struct {
	cfg         *config.Config
	screen      screen
	cursor      int
	accountInfo string
	debugMode   bool
	menuItems   []menuItem

	// text inputs for forms
	inputs     []textinput.Model
	inputFocus int

	// loading
	spinner spinner.Model
	loading bool

	// results
	resultTitle string
	resultLines []string

	// debug log viewer
	logViewport viewport.Model
	logContent  string
	copyStatus  string

	// status message
	statusMsg string

	width  int
	height int
}

func NewModel(cfg *config.Config, debugMode bool) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	items := []menuItem{
		{title: "🔑  Login com Stripe", desc: "Autenticar via stripe login (abre o navegador)"},
		{title: "📦  Seed: Produtos", desc: "Criar produtos com nomes aleatórios"},
		{title: "💰  Seed: Produtos + Preços", desc: "Criar produtos com preços aleatórios"},
		{title: "👤  Seed: Clientes", desc: "Criar clientes com dados aleatórios"},
		{title: "💳  Seed: Payment Intents", desc: "Criar pagamentos confirmados com cartão de teste"},
	}
	if debugMode {
		items = append(items, menuItem{
			title: "🐛  Ver Log de Debug",
			desc:  "Visualizar e copiar o log de debug",
		})
	}
	items = append(items, menuItem{title: "🚪  Sair", desc: "Encerrar o programa"})

	return Model{
		cfg:       cfg,
		screen:    screenMain,
		spinner:   s,
		debugMode: debugMode,
		menuItems: items,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		info, err := stripeClient.CheckLogin()
		return checkLoginMsg{info: info, err: err}
	})
}

// --- Update ------------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logViewport.Width = msg.Width - 4
		m.logViewport.Height = msg.Height - 8
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
		logger.Log("seed done [%s]: created=%d errors=%d", msg.label, msg.result.Created, msg.result.Errors)
		return m, nil

	case checkLoginMsg:
		m.loading = false
		if msg.err != nil {
			logger.Log("checkLogin failed: %v", msg.err)
			m.accountInfo = ""
			m.statusMsg = warningStyle.Render("⚠ " + msg.err.Error())
		} else {
			logger.Log("checkLogin ok: %s", msg.info)
			m.accountInfo = msg.info
			m.statusMsg = ""
		}
		return m, nil

	case loginDoneMsg:
		m.screen = screenMain
		if msg.err != nil {
			logger.Log("stripe login failed: %v", msg.err)
			m.statusMsg = errorStyle.Render("✗ Falha no login: " + msg.err.Error())
			return m, nil
		}
		logger.Log("stripe login completed successfully")
		m.loading = true
		m.statusMsg = "Verificando autenticação..."
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			info, err := stripeClient.CheckLogin()
			return checkLoginMsg{info: info, err: err}
		})

	case loadLogMsg:
		m.logContent = msg.content
		m.logViewport.SetContent(msg.content)
		m.logViewport.GotoBottom()
		return m, nil
	}

	switch m.screen {
	case screenMain:
		return m.updateMain(msg)
	case screenLogin:
		return m.updateLogin(msg)
	case screenSeedProducts:
		return m.updateSeedProducts(msg)
	case screenSeedProductsPrices:
		return m.updateSeedProductsPrices(msg)
	case screenSeedCustomers:
		return m.updateSeedCustomers(msg)
	case screenSeedPaymentIntents:
		return m.updateSeedPaymentIntents(msg)
	case screenResults:
		return m.updateResults(msg)
	case screenDebugLog:
		return m.updateDebugLog(msg)
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
			if m.cursor < len(m.menuItems)-1 {
				m.cursor++
			}
		case msg.String() == "enter":
			return m.handleMenuSelect()
		}
	}
	return m, nil
}

func (m Model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.screen = screenMain
			return m, nil
		case "enter":
			return m, tea.ExecProcess(stripeClient.LoginCmd(), func(err error) tea.Msg {
				return loginDoneMsg{err: err}
			})
		}
	}
	return m, nil
}

func (m Model) handleMenuSelect() (tea.Model, tea.Cmd) {
	quitIdx := len(m.menuItems) - 1
	debugLogIdx := quitIdx - 1 // only valid when debugMode

	switch m.cursor {
	case 0: // login com stripe
		m.screen = screenLogin
		m.statusMsg = ""
		return m, nil

	case 1: // seed products
		if m.accountInfo == "" {
			m.statusMsg = warningStyle.Render("⚠ Faça login primeiro com 'Login com Stripe'.")
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

	case 2: // seed products + prices
		if m.accountInfo == "" {
			m.statusMsg = warningStyle.Render("⚠ Faça login primeiro com 'Login com Stripe'.")
			return m, nil
		}
		m.screen = screenSeedProductsPrices
		m.inputs = make([]textinput.Model, 4)
		fields := []struct {
			placeholder string
		}{{"10"}, {"500"}, {"50000"}, {"brl"}}
		for i, f := range fields {
			ti := textinput.New()
			ti.Placeholder = f.placeholder
			ti.CharLimit = 20
			ti.Width = 20
			m.inputs[i] = ti
		}
		m.inputs[0].Focus()
		m.inputFocus = 0
		m.statusMsg = ""
		return m, textinput.Blink

	case 3: // seed customers
		if m.accountInfo == "" {
			m.statusMsg = warningStyle.Render("⚠ Faça login primeiro com 'Login com Stripe'.")
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

	case 4: // seed payment intents
		if m.accountInfo == "" {
			m.statusMsg = warningStyle.Render("⚠ Faça login primeiro com 'Login com Stripe'.")
			return m, nil
		}
		m.screen = screenSeedPaymentIntents
		m.inputs = make([]textinput.Model, 5)
		piFields := []struct {
			placeholder string
			charLimit   int
		}{
			{"10", 5},        // quantity
			{"10.00", 15},    // min amount (currency units)
			{"200.00", 15},   // max amount (currency units)
			{"brl", 10},      // currency
			{"pm_card_visa", 40}, // payment method
		}
		for i, f := range piFields {
			ti := textinput.New()
			ti.Placeholder = f.placeholder
			ti.CharLimit = f.charLimit
			ti.Width = 40
			m.inputs[i] = ti
		}
		m.inputs[0].Focus()
		m.inputFocus = 0
		m.statusMsg = ""
		return m, textinput.Blink

	default:
		if m.debugMode && m.cursor == debugLogIdx {
			vp := viewport.New(m.width-4, m.height-8)
			m.logViewport = vp
			m.copyStatus = ""
			m.screen = screenDebugLog
			return m, func() tea.Msg {
				return loadLogMsg{content: readLogFile()}
			}
		}
		if m.cursor == quitIdx {
			return m, tea.Quit
		}
	}
	return m, nil
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
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				res := stripeClient.SeedProducts(n)
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
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				res := stripeClient.SeedProductsWithPrices(n, minC, maxC, cur)
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
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				res := stripeClient.SeedCustomers(n)
				return seedDoneMsg{result: res, label: "Seed: Clientes"}
			})
		}
	}
	var cmd tea.Cmd
	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func (m Model) updateSeedPaymentIntents(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			minAmt := parseFloatOr(m.inputs[1].Value(), 10.00)
			maxAmt := parseFloatOr(m.inputs[2].Value(), 200.00)
			cur := strings.ToLower(strings.TrimSpace(m.inputs[3].Value()))
			if cur == "" {
				cur = "brl"
			}
			pm := strings.TrimSpace(m.inputs[4].Value())
			if pm == "" {
				pm = "pm_card_visa"
			}
			if minAmt > maxAmt {
				minAmt, maxAmt = maxAmt, minAmt
			}

			m.screen = screenLoading
			m.loading = true
			m.statusMsg = fmt.Sprintf("Criando %d payment intents...", n)
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				res := stripeClient.SeedPaymentIntents(n, minAmt, maxAmt, cur, pm)
				return seedDoneMsg{result: res, label: "Seed: Payment Intents"}
			})
		}
	}
	var cmd tea.Cmd
	m.inputs[m.inputFocus], cmd = m.inputs[m.inputFocus].Update(msg)
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

func (m Model) updateDebugLog(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.screen = screenMain
			m.copyStatus = ""
			return m, nil
		case "c":
			if m.logContent == "" {
				m.copyStatus = warningStyle.Render("⚠ Log vazio")
			} else if err := clipboard.WriteAll(m.logContent); err != nil {
				logger.Log("clipboard copy failed: %v", err)
				m.copyStatus = errorStyle.Render("✗ Erro ao copiar: " + err.Error())
			} else {
				logger.Log("log copied to clipboard")
				m.copyStatus = successStyle.Render("✓ Log copiado!")
			}
			return m, nil
		case "r":
			return m, func() tea.Msg {
				return loadLogMsg{content: readLogFile()}
			}
		}
	}
	var cmd tea.Cmd
	m.logViewport, cmd = m.logViewport.Update(msg)
	return m, cmd
}

// --- Helpers -----------------------------------------------------------------

func readLogFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "(erro ao localizar home)"
	}
	path := filepath.Join(home, ".stripe-seeder-debug.log")
	data, err := os.ReadFile(path)
	if err != nil {
		return "(log não encontrado — nenhuma ação registrada ainda)"
	}
	return string(data)
}

func parseFloatOr(s string, fallback float64) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
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
