# 🌱 Stripe Seeder TUI

Uma aplicação TUI (Terminal User Interface) feita com [Charm](https://charm.land/) para semear dados no Stripe rapidamente.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Stripe](https://img.shields.io/badge/Stripe-API-635BFF?logo=stripe)
![Charm](https://img.shields.io/badge/Charm-Bubble%20Tea-FF69B4)

## ✨ Funcionalidades

- **Ver Projeto Conectado** — exibe chave mascarada e ambiente (test/live)
- **Alterar Chave API** — configura e valida a secret key em tempo real
- **Seed: Produtos** — cria N produtos com nomes e descrições aleatórias
- **Seed: Produtos + Preços** — cria produtos com preço vinculado (valor, moeda, faixa configuráveis)
- **Seed: Clientes** — cria clientes com nomes BR e emails fake
- **Persistência** — salva configuração em `~/.stripe-seeder.json`

## 📁 Estrutura

```
stripe-seeder/
├── go.mod
├── main.go
└── internal/
    ├── config/
    │   └── config.go          # Persistência da configuração
    ├── stripe/
    │   └── client.go          # Operações com a API do Stripe
    └── tui/
        ├── keys.go            # Key bindings
        ├── model.go           # Bubble Tea model (lógica)
        ├── styles.go          # Lip Gloss estilos
        └── views.go           # Renderização das telas
```

## 🚀 Como usar

### Pré-requisitos

- Go 1.21+
- Uma Stripe Secret Key (`sk_test_...` ou `sk_live_...`)

### Instalação

```bash
git clone <seu-repo>
cd stripe-seeder

# Baixar dependências
go mod tidy

# Rodar
go run main.go
```

### Build

```bash
go build -o stripe-seeder .
./stripe-seeder
```

## 🎮 Controles

| Tecla       | Ação               |
|-------------|---------------------|
| `↑` / `k`  | Navegar para cima   |
| `↓` / `j`  | Navegar para baixo  |
| `Enter`     | Selecionar / Confirmar |
| `Tab`       | Próximo campo       |
| `Esc`       | Voltar              |
| `q`         | Sair                |

## ⚠️ Aviso

- **Use SEMPRE em ambiente de teste** (`sk_test_...`).
- Todos os itens criados possuem metadata `seeder: stripe-seeder` para fácil identificação e limpeza.

## 📦 Dependências

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — Framework TUI
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Estilos
- [Bubbles](https://github.com/charmbracelet/bubbles) — Componentes (input, spinner)
- [stripe-go](https://github.com/stripe/stripe-go) — SDK oficial do Stripe
