package generator

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"importer/models"
)

type GenerationSummary struct {
	CustomerCount int
	AccountCount  int
	LinkCount     int
}

type DataGenerator struct {
	config  GeneratorConfig
	names   NameGenerator
	summary GenerationSummary
}

type GeneratorConfig struct {
	NumCustomers    int
	MultiAcctChance float32
	ThirdAcctChance float32
	CustomerPrefix  string
	AccountPrefix   string
}

type NameGenerator struct {
	firstNames []string
	lastNames  []string
}

func (g *DataGenerator) GetSummary() GenerationSummary {
	return g.summary
}

func NewGenerator(cfg GeneratorConfig) *DataGenerator {
	return &DataGenerator{
		config: cfg,
		names: NameGenerator{
			firstNames: []string{"John", "Jane", "Michael", "Sarah", "David", "Lisa", "Robert", "Emily", "William", "Emma"},
			lastNames:  []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"},
		},
	}
}

func (g *DataGenerator) generateName() string {
	firstName := g.names.firstNames[rand.Intn(len(g.names.firstNames))]
	lastName := g.names.lastNames[rand.Intn(len(g.names.lastNames))]
	return firstName + " " + lastName
}

func (g *DataGenerator) GenerateCustomers() []models.Customer {
	customers := make([]models.Customer, g.config.NumCustomers)

	log.Println("Generating customer data...")
	for i := 0; i < g.config.NumCustomers; i++ {
		customerNum := fmt.Sprintf("%s%06d", g.config.CustomerPrefix, i+1)

		customers[i] = models.Customer{
			ClientID:       fmt.Sprintf("CLI%06d", i+1),
			CustomerNumber: customerNum,
			CustomerName:   fmt.Sprintf("Customer %d Corp", i+1),
			Address:        fmt.Sprintf("%d Business St, Suite %d", rand.Intn(999)+1, rand.Intn(100)+1),
			Name:           g.generateName(),
			Email:          fmt.Sprintf("contact%d@customer%d.com", i+1, i+1),
		}

		if (i+1)%10000 == 0 {
			log.Printf("Generated %d customers...", i+1)
		}
	}

	g.summary.CustomerCount = len(customers)
	return customers
}

func (g *DataGenerator) GenerateAccounts() []models.Account {
	accounts := make([]models.Account, g.config.NumCustomers)

	log.Println("Generating account data...")
	for i := 0; i < g.config.NumCustomers; i++ {
		accounts[i] = models.Account{
			AccountNumber: fmt.Sprintf("%s%06d", g.config.AccountPrefix, i+1),
			AccountName:   fmt.Sprintf("Account %d", i+1),
		}

		if (i+1)%10000 == 0 {
			log.Printf("Generated %d accounts...", i+1)
		}
	}

	g.summary.AccountCount = len(accounts)
	return accounts
}

func (g *DataGenerator) GenerateLinks() []models.CustomerAccount {
	// Pre-allocate a slice with estimated capacity
	estimatedLinks := g.config.NumCustomers * 2
	links := make([]models.CustomerAccount, 0, estimatedLinks)

	log.Println("Generating customer-account links...")

	// First, ensure each customer has their primary account
	for i := 0; i < g.config.NumCustomers; i++ {
		custNum := i + 1
		links = append(links, models.CustomerAccount{
			CustomerNumber: fmt.Sprintf("%s%06d", g.config.CustomerPrefix, custNum),
			AccountNumber:  fmt.Sprintf("%s%06d", g.config.AccountPrefix, custNum),
		})

		// 30% chance of getting a second account
		if rand.Float32() < g.config.MultiAcctChance {
			// Pick a random account from the pool
			extraAccount := rand.Intn(g.config.NumCustomers) + 1
			if extraAccount != custNum { // Avoid duplicate links
				links = append(links, models.CustomerAccount{
					CustomerNumber: fmt.Sprintf("%s%06d", g.config.CustomerPrefix, custNum),
					AccountNumber:  fmt.Sprintf("%s%06d", g.config.AccountPrefix, extraAccount),
				})

				// 10% chance of getting a third account
				if rand.Float32() < g.config.ThirdAcctChance {
					extraAccount = rand.Intn(g.config.NumCustomers) + 1
					if extraAccount != custNum { // Avoid duplicate links
						links = append(links, models.CustomerAccount{
							CustomerNumber: fmt.Sprintf("%s%06d", g.config.CustomerPrefix, custNum),
							AccountNumber:  fmt.Sprintf("%s%06d", g.config.AccountPrefix, extraAccount),
						})
					}
				}
			}
		}

		if custNum%10000 == 0 {
			log.Printf("Generated links for %d customers...", custNum)
		}
	}

	log.Printf("Generated %d total links", len(links))

	g.summary.LinkCount = len(links)
	return links
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
