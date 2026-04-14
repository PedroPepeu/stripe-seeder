package stripe

import (
	"fmt"
	"math/rand"
	"time"

	stripelib "github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/price"
	"github.com/stripe/stripe-go/v78/product"
)

var rng *rand.Rand

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// ValidateKey checks if the Stripe key is functional by listing 1 product.
func ValidateKey(apiKey string) (string, error) {
	stripelib.Key = apiKey

	params := &stripelib.ProductListParams{}
	params.Limit = stripelib.Int64(1)
	iter := product.List(params)
	// just attempt to iterate — if key is bad, this returns an error
	iter.Next()
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("chave inválida: %w", err)
	}

	env := "LIVE 🔴"
	if len(apiKey) > 3 && apiKey[:7] == "sk_test" {
		env = "TEST 🟡"
	}

	return env, nil
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

// SeedProducts creates n products with random names and returns results.
func SeedProducts(apiKey string, n int) SeedResult {
	stripelib.Key = apiKey
	res := SeedResult{}

	for i := 0; i < n; i++ {
		name := randomProductName()
		params := &stripelib.ProductParams{
			Name:        stripelib.String(name),
			Description: stripelib.String(randomDescription()),
			Metadata: map[string]string{
				"seeder": "stripe-seeder",
				"batch":  fmt.Sprintf("%d", time.Now().Unix()),
			},
		}

		p, err := product.New(params)
		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✗ Erro: %s", err.Error()))
		} else {
			res.Created++
			res.Details = append(res.Details, fmt.Sprintf("✓ Produto: %s (id: %s)", p.Name, p.ID))
		}
	}
	return res
}

// SeedProductsWithPrices creates n products each with a random price attached.
func SeedProductsWithPrices(apiKey string, n int, minCents, maxCents int64, currency string) SeedResult {
	stripelib.Key = apiKey
	res := SeedResult{}

	for i := 0; i < n; i++ {
		name := randomProductName()
		prodParams := &stripelib.ProductParams{
			Name:        stripelib.String(name),
			Description: stripelib.String(randomDescription()),
			Metadata: map[string]string{
				"seeder": "stripe-seeder",
				"batch":  fmt.Sprintf("%d", time.Now().Unix()),
			},
		}

		p, err := product.New(prodParams)
		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✗ Produto erro: %s", err.Error()))
			continue
		}

		amount := randomPrice(minCents, maxCents)
		priceParams := &stripelib.PriceParams{
			Product:    stripelib.String(p.ID),
			UnitAmount: stripelib.Int64(amount),
			Currency:   stripelib.String(currency),
		}

		pr, err := price.New(priceParams)
		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✓ Produto: %s | ✗ Preço erro: %s", p.ID, err.Error()))
		} else {
			res.Created++
			res.Details = append(res.Details, fmt.Sprintf("✓ %s → %s %.2f (price: %s)", name, currency, float64(pr.UnitAmount)/100, pr.ID))
		}
	}
	return res
}

// SeedCustomers creates n customers with random data.
func SeedCustomers(apiKey string, n int) SeedResult {
	stripelib.Key = apiKey
	res := SeedResult{}

	for i := 0; i < n; i++ {
		first := firstNames[rng.Intn(len(firstNames))]
		last := lastNames[rng.Intn(len(lastNames))]
		fullName := fmt.Sprintf("%s %s", first, last)
		email := randomEmail(first, last)

		params := &stripelib.CustomerParams{
			Name:  stripelib.String(fullName),
			Email: stripelib.String(email),
			Metadata: map[string]string{
				"seeder": "stripe-seeder",
				"batch":  fmt.Sprintf("%d", time.Now().Unix()),
			},
		}

		c, err := customer.New(params)
		if err != nil {
			res.Errors++
			res.Details = append(res.Details, fmt.Sprintf("✗ Erro: %s", err.Error()))
		} else {
			res.Created++
			res.Details = append(res.Details, fmt.Sprintf("✓ %s <%s> (id: %s)", c.Name, c.Email, c.ID))
		}
	}
	return res
}
