package stripe

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/PedroPepeu/stripe-seeder/internal/logger"
)

var rng *rand.Rand

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// CheckLogin runs stripe whoami to verify authentication status.
func CheckLogin() (string, error) {
	logger.Log("cmd: stripe whoami")
	out, err := exec.Command("stripe", "whoami").CombinedOutput()
	logger.Log("stripe whoami output: %s", strings.TrimSpace(string(out)))
	if err != nil {
		logger.Log("stripe whoami error: %v", err)
		return "", fmt.Errorf("não autenticado. Use 'Login com Stripe'")
	}
	return strings.TrimSpace(string(out)), nil
}

// LoginCmd returns a *exec.Cmd for stripe login (run via tea.ExecProcess).
func LoginCmd() *exec.Cmd {
	logger.Log("cmd: stripe login")
	return exec.Command("stripe", "login")
}

// --- Random data generators ---------------------------------------------------

var productAdjectives = []string{
	"Premium", "Ultra", "Mega", "Pro", "Elite", "Turbo", "Hyper",
	"Smart", "Eco", "Nano", "Quantum", "Stellar", "Cosmic", "Zen",
	"Prime", "Royal", "Golden", "Crystal", "Titan", "Nova",
}

var productNouns = []string{
	"Widget", "Gadget", "Module", "Pack", "Kit", "Bundle", "Suite",
	"Service", "Plan", "License", "Token", "Credit", "Pass", "Boost",
	"Shield", "Vault", "Engine", "Matrix", "Nexus", "Core",
}

var firstNames = []string{
	"Alice", "Bruno", "Carla", "Diego", "Elena", "Felipe", "Gabi",
	"Hugo", "Iris", "João", "Karen", "Lucas", "Maria", "Nina",
	"Oscar", "Paula", "Rafael", "Sofia", "Thiago", "Valentina",
}

var lastNames = []string{
	"Silva", "Santos", "Oliveira", "Souza", "Rodrigues", "Ferreira",
	"Almeida", "Pereira", "Lima", "Costa", "Ribeiro", "Martins",
	"Carvalho", "Gomes", "Rocha", "Araújo", "Barbosa", "Melo",
}

var domains = []string{
	"example.com", "test.io", "demo.dev", "seed.app", "faker.org",
}

func randomProductName() string {
	adj := productAdjectives[rng.Intn(len(productAdjectives))]
	noun := productNouns[rng.Intn(len(productNouns))]
	return fmt.Sprintf("%s %s", adj, noun)
}

func randomDescription() string {
	descs := []string{
		"Solução completa para seu negócio.",
		"O melhor custo-benefício do mercado.",
		"Tecnologia de ponta para você.",
		"Automatize processos com facilidade.",
		"Potencialize seus resultados.",
		"Item gerado via stripe-seeder.",
		"Produto de demonstração para testes.",
		"Perfeito para ambientes de staging.",
	}
	return descs[rng.Intn(len(descs))]
}

func randomEmail(first, last string) string {
	domain := domains[rng.Intn(len(domains))]
	num := rng.Intn(999)
	return fmt.Sprintf("%s.%s%d@%s", first, last, num, domain)
}

// randomPrice returns a value in cents between min and max (inclusive).
func randomPrice(minCents, maxCents int64) int64 {
	return minCents + rng.Int63n(maxCents-minCents+1)
}

// --- Seed operations ----------------------------------------------------------

type SeedResult struct {
	Created int
	Errors  int
	Details []string
}

// parseID extracts the "id" field from stripe-cli JSON output.
func parseID(data []byte) string {
	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		logger.Log("parseID error: %v | raw: %s", err, strings.TrimSpace(string(data)))
		return "?"
	}
	return obj.ID
}

// runStripe executes a stripe-cli command and returns combined output.
func runStripe(args ...string) ([]byte, error) {
	logger.Log("cmd: stripe %s", strings.Join(args, " "))
	out, err := exec.Command("stripe", args...).CombinedOutput()
	if err != nil {
		logger.Log("error: %v | output: %s", err, strings.TrimSpace(string(out)))
	} else {
		logger.Log("output: %s", strings.TrimSpace(string(out)))
	}
	return out, err
}

// SeedProducts creates n products via stripe-cli.
func SeedProducts(n int) SeedResult {
	res := SeedResult{}
	batch := fmt.Sprintf("%d", time.Now().Unix())
	logger.Log("SeedProducts: n=%d batch=%s", n, batch)

	for i := 0; i < n; i++ {
		name := randomProductName()
		out, err := runStripe("products", "create",
			"-d", "name="+name,
			"-d", "description="+randomDescription(),
			"-d", "metadata[seeder]=stripe-seeder",
			"-d", "metadata[batch]="+batch,
		)

		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✗ Produto erro: %s", strings.TrimSpace(string(out))))
		} else {
			id := parseID(out)
			res.Created++
			res.Details = append(res.Details, fmt.Sprintf("✓ Produto: %s (id: %s)", name, id))
		}
	}
	logger.Log("SeedProducts done: created=%d errors=%d", res.Created, res.Errors)
	return res
}

// SeedProductsWithPrices creates n products each with a random price via stripe-cli.
func SeedProductsWithPrices(n int, minCents, maxCents int64, currency string) SeedResult {
	res := SeedResult{}
	batch := fmt.Sprintf("%d", time.Now().Unix())
	logger.Log("SeedProductsWithPrices: n=%d min=%d max=%d currency=%s batch=%s", n, minCents, maxCents, currency, batch)

	for i := 0; i < n; i++ {
		name := randomProductName()
		prodOut, err := runStripe("products", "create",
			"-d", "name="+name,
			"-d", "description="+randomDescription(),
			"-d", "metadata[seeder]=stripe-seeder",
			"-d", "metadata[batch]="+batch,
		)

		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✗ Produto erro: %s", strings.TrimSpace(string(prodOut))))
			continue
		}

		prodID := parseID(prodOut)
		amount := randomPrice(minCents, maxCents)

		priceOut, err := runStripe("prices", "create",
			"-d", "product="+prodID,
			"-d", fmt.Sprintf("unit_amount=%d", amount),
			"-d", "currency="+currency,
		)

		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✓ Produto: %s | ✗ Preço erro: %s", prodID, strings.TrimSpace(string(priceOut))))
		} else {
			priceID := parseID(priceOut)
			res.Created++
			res.Details = append(res.Details, fmt.Sprintf("✓ %s → %s %.2f (price: %s)", name, currency, float64(amount)/100, priceID))
		}
	}
	logger.Log("SeedProductsWithPrices done: created=%d errors=%d", res.Created, res.Errors)
	return res
}

// SeedPaymentIntents creates n confirmed payment intents via stripe-cli.
// minAmount and maxAmount are in currency units (e.g. 50.00 for R$50,00) — converted to cents internally.
func SeedPaymentIntents(n int, minAmount, maxAmount float64, currency, paymentMethod string) SeedResult {
	res := SeedResult{}
	batch := fmt.Sprintf("%d", time.Now().Unix())
	minCents := int64(minAmount * 100)
	maxCents := int64(maxAmount * 100)
	logger.Log("SeedPaymentIntents: n=%d min=%.2f max=%.2f currency=%s pm=%s batch=%s",
		n, minAmount, maxAmount, currency, paymentMethod, batch)

	for i := 0; i < n; i++ {
		amount := randomPrice(minCents, maxCents)

		out, err := runStripe("payment_intents", "create",
			"-d", fmt.Sprintf("amount=%d", amount),
			"-d", "currency="+currency,
			"-d", "payment_method="+paymentMethod,
			"-d", "payment_method_types[]=card",
			"-d", "confirm=true",
			"-d", "metadata[seeder]=stripe-seeder",
			"-d", "metadata[batch]="+batch,
		)

		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✗ Erro: %s", strings.TrimSpace(string(out))))
		} else {
			id := parseID(out)
			res.Created++
			res.Details = append(res.Details, fmt.Sprintf("✓ PaymentIntent: %.2f %s (id: %s)",
				float64(amount)/100, strings.ToUpper(currency), id))
		}
	}
	logger.Log("SeedPaymentIntents done: created=%d errors=%d", res.Created, res.Errors)
	return res
}

// SeedCustomers creates n customers via stripe-cli.
func SeedCustomers(n int) SeedResult {
	res := SeedResult{}
	batch := fmt.Sprintf("%d", time.Now().Unix())
	logger.Log("SeedCustomers: n=%d batch=%s", n, batch)

	for i := 0; i < n; i++ {
		first := firstNames[rng.Intn(len(firstNames))]
		last := lastNames[rng.Intn(len(lastNames))]
		fullName := fmt.Sprintf("%s %s", first, last)
		email := randomEmail(first, last)

		out, err := runStripe("customers", "create",
			"-d", "name="+fullName,
			"-d", "email="+email,
			"-d", "metadata[seeder]=stripe-seeder",
			"-d", "metadata[batch]="+batch,
		)

		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✗ Erro: %s", strings.TrimSpace(string(out))))
		} else {
			id := parseID(out)
			res.Created++
			res.Details = append(res.Details, fmt.Sprintf("✓ %s <%s> (id: %s)", fullName, email, id))
		}
	}
	logger.Log("SeedCustomers done: created=%d errors=%d", res.Created, res.Errors)
	return res
}
