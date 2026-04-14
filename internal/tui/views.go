package tui

import (
	"fmt"
	"strings"
)

const logo = `
 _____ _        _               _____               _           
/  ___| |      (_)             /  ___|             | |          
\ ` + "`" + `--.| |_ _ __ _ _ __   ___  \ ` + "`" + `--.  ___  ___  __| | ___ _ __ 
 ` + "`" + `--. \ __| '__| | '_ \ / _ \  ` + "`" + `--. \/ _ \/ _ \/ _` + "`" + ` |/ _ \ '__|
/\__/ / |_| |  | | |_) |  __/ /\__/ /  __/  __/ (_| |  __/ |   
\____/ \__|_|  |_| .__/ \___| \____/ \___|\___|\__,_|\___|_|   
                  | |                                            
                  |_|                                            `

func (m Model) View() string {
	switch m.screen {
	case screenMain:
		return m.viewMain()
	case screenSetKey:
		return m.viewSetKey()
	case screenSeedProducts:
		return m.viewSeedProducts()
	case screenSeedProductsPrices:
		return m.viewSeedProductsPrices()
	case screenSeedCustomers:
		return m.viewSeedCustomers()
	case screenResults:
		return m.viewResults()
	case screenLoading:
		return m.viewLoading()
	}
	return ""
}

func (m Model) viewMain() string {
	var b strings.Builder

	b.WriteString(logoStyle.Render(logo))
	b.WriteString("\n")

	// Connection status bar
	if m.cfg.APIKey != "" {
		projName := m.cfg.ProjectName
		if projName == "" {
			projName = "Sem nome"
		}
		status := fmt.Sprintf("  Projeto: %s  │  Ambiente: %s  │  Chave: %s  ",
			projName, m.env, maskKey(m.cfg.APIKey))
		b.WriteString(statusBoxStyle.Render(status))
	} else {
		b.WriteString(statusBoxStyle.Render("  ⚠ Nenhuma chave API configurada  "))
	}
	b.WriteString("\n\n")

	b.WriteString(titleStyle.Render("  MENU PRINCIPAL"))
	b.WriteString("\n\n")

	for i, item := range mainMenuItems {
		cursor := "  "
		style := menuItemStyle
		if i == m.cursor {
			cursor = "▸ "
			style = selectedMenuStyle
		}
		b.WriteString(style.Render(cursor + item.title))
		b.WriteString("\n")
		if i == m.cursor {
			b.WriteString(descriptionStyle.Render(item.desc))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if m.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(statusBoxStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("  ↑/↓ navegar • enter selecionar • q sair"))

	return b.String()
}

func (m Model) viewSetKey() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  🔑 CONFIGURAR CHAVE API"))
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("  Stripe Secret Key:"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("  Nome do Projeto (opcional):"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[1].View())
	b.WriteString("\n\n")

	if m.statusMsg != "" {
		b.WriteString("  " + m.statusMsg)
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("  tab trocar campo • enter salvar e validar • esc voltar"))

	return b.String()
}

func (m Model) viewSeedProducts() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  📦 SEED: PRODUTOS"))
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render("  Cria produtos com nomes e descrições aleatórias no Stripe."))
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("  Quantidade de produtos:"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("  enter executar • esc voltar"))

	return b.String()
}

func (m Model) viewSeedProductsPrices() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  💰 SEED: PRODUTOS + PREÇOS"))
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render("  Cria produtos + preço vinculado com valor aleatório."))
	b.WriteString("\n\n")

	labels := []string{
		"Quantidade:",
		"Preço mínimo (centavos):",
		"Preço máximo (centavos):",
		"Moeda (ex: brl, usd):",
	}

	for i, label := range labels {
		b.WriteString(inputLabelStyle.Render("  " + label))
		b.WriteString("\n")
		b.WriteString("  " + m.inputs[i].View())
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("  tab próximo campo • enter executar • esc voltar"))

	return b.String()
}

func (m Model) viewSeedCustomers() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  👤 SEED: CLIENTES"))
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render("  Cria clientes com nomes e emails aleatórios."))
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("  Quantidade de clientes:"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("  enter executar • esc voltar"))

	return b.String()
}

func (m Model) viewResults() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("  ✅ RESULTADO: %s", m.resultTitle)))
	b.WriteString("\n")

	if m.statusMsg != "" {
		b.WriteString("  " + m.statusMsg)
		b.WriteString("\n")
	}

	if len(m.resultLines) > 0 {
		content := strings.Join(m.resultLines, "\n")
		b.WriteString(resultBoxStyle.Render(content))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  enter/esc voltar ao menu"))

	return b.String()
}

func (m Model) viewLoading() string {
	var b strings.Builder

	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s %s", m.spinner.View(), m.statusMsg))
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render("  Aguarde, comunicando com a API do Stripe..."))

	return b.String()
}
