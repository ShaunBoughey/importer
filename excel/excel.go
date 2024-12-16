package excel

import (
	"fmt"
	"log"
	"time"

	"importer/generator"
	"importer/models"

	"github.com/xuri/excelize/v2"
)

type Importer struct {
	db  models.CustomerRepository
	cfg interface{}
}

func NewImporter(db models.CustomerRepository, cfg interface{}) *Importer {
	return &Importer{
		db:  db,
		cfg: cfg,
	}
}

// GenerateFile creates a new Excel file with generated data
func GenerateFile(filename string, gen *generator.DataGenerator) error {
	f := excelize.NewFile()
	defer f.Close()

	// Generate the data
	customers := gen.GenerateCustomers()
	accounts := gen.GenerateAccounts()
	links := gen.GenerateLinks()

	// Create customers sheet
	customerSheet := "Customers"
	f.SetSheetName("Sheet1", customerSheet)

	// Set headers for Customers
	headers := []string{"Client ID", "Customer Number", "Customer Name", "Address", "Name", "Email"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(customerSheet, cell, header)
	}

	// Write customer data
	for i, customer := range customers {
		row := i + 2
		f.SetCellValue(customerSheet, fmt.Sprintf("A%d", row), customer.ClientID)
		f.SetCellValue(customerSheet, fmt.Sprintf("B%d", row), customer.CustomerNumber)
		f.SetCellValue(customerSheet, fmt.Sprintf("C%d", row), customer.CustomerName)
		f.SetCellValue(customerSheet, fmt.Sprintf("D%d", row), customer.Address)
		f.SetCellValue(customerSheet, fmt.Sprintf("E%d", row), customer.Name)
		f.SetCellValue(customerSheet, fmt.Sprintf("F%d", row), customer.Email)
	}

	// Create accounts sheet
	accountSheet := "Account"
	f.NewSheet(accountSheet)

	// Set headers for Accounts
	f.SetCellValue(accountSheet, "A1", "Account Number")
	f.SetCellValue(accountSheet, "B1", "Account Name")

	// Write account data
	for i, account := range accounts {
		row := i + 2
		f.SetCellValue(accountSheet, fmt.Sprintf("A%d", row), account.AccountNumber)
		f.SetCellValue(accountSheet, fmt.Sprintf("B%d", row), account.AccountName)
	}

	// Create customer account links sheet
	linkSheet := "customer account link"
	f.NewSheet(linkSheet)

	// Set headers for Links
	f.SetCellValue(linkSheet, "A1", "Customer Number")
	f.SetCellValue(linkSheet, "B1", "Account Number")

	// Write link data
	for i, link := range links {
		row := i + 2
		f.SetCellValue(linkSheet, fmt.Sprintf("A%d", row), link.CustomerNumber)
		f.SetCellValue(linkSheet, fmt.Sprintf("B%d", row), link.AccountNumber)
	}

	// Save the file
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save Excel file: %v", err)
	}

	// Get and print summary
	summary := gen.GetSummary()
	log.Printf("\nGeneration Summary:")
	log.Printf("------------------")
	log.Printf("Total Customers: %d", summary.CustomerCount)
	log.Printf("Total Accounts:  %d", summary.AccountCount)
	log.Printf("Total Links:     %d", summary.LinkCount)
	log.Printf("File generated successfully: %s", filename)

	return nil
}

// Import reads an Excel file and imports the data
func (imp *Importer) Import(filename string) error {
	start := time.Now()
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %v", err)
	}
	defer f.Close()

	// Read all data first
	customers, err := readCustomers(f)
	if err != nil {
		return fmt.Errorf("failed to read customers: %v", err)
	}
	log.Printf("Read %d customers from file", len(customers))

	accounts, err := readAccounts(f)
	if err != nil {
		return fmt.Errorf("failed to read accounts: %v", err)
	}
	log.Printf("Read %d accounts from file", len(accounts))

	links, err := readLinks(f)
	if err != nil {
		return fmt.Errorf("failed to read customer-account links: %v", err)
	}
	log.Printf("Read %d links from file", len(links))

	// Insert customers
	log.Printf("Inserting customers...")
	customerIDs, err := imp.db.InsertCustomers(customers)
	if err != nil {
		return fmt.Errorf("failed to insert customers: %v", err)
	}
	log.Printf("Inserted/Updated %d customers", len(customerIDs))

	// Insert accounts
	// Note: We need to cast the interface to use AccountRepository methods
	if accountRepo, ok := imp.db.(models.AccountRepository); ok {
		log.Printf("Inserting accounts...")
		accountIDs, err := accountRepo.InsertAccounts(accounts)
		if err != nil {
			return fmt.Errorf("failed to insert accounts: %v", err)
		}
		log.Printf("Inserted/Updated %d accounts", len(accountIDs))

		// Insert customer-account links
		if linkRepo, ok := imp.db.(models.CustomerAccountRepository); ok {
			log.Printf("Inserting customer-account links...")
			err = linkRepo.InsertCustomerAccounts(links, customerIDs, accountIDs)
			if err != nil {
				return fmt.Errorf("failed to insert customer-account links: %v", err)
			}
		}
	}

	log.Printf("Import completed successfully in %v", time.Since(start))
	return nil
}
func readCustomers(f *excelize.File) ([]models.Customer, error) {
	rows, err := f.GetRows("Customers")
	if err != nil {
		return nil, err
	}

	var customers []models.Customer
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}
		if len(row) < 6 {
			continue // Skip incomplete rows
		}
		customers = append(customers, models.Customer{
			ClientID:       row[0],
			CustomerNumber: row[1],
			CustomerName:   row[2],
			Address:        row[3],
			Name:           row[4],
			Email:          row[5],
		})
	}
	return customers, nil
}

func readAccounts(f *excelize.File) ([]models.Account, error) {
	rows, err := f.GetRows("Account")
	if err != nil {
		return nil, err
	}

	var accounts []models.Account
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}
		if len(row) < 2 {
			continue // Skip incomplete rows
		}
		accounts = append(accounts, models.Account{
			AccountNumber: row[0],
			AccountName:   row[1],
		})
	}
	return accounts, nil
}

func readLinks(f *excelize.File) ([]models.CustomerAccount, error) {
	rows, err := f.GetRows("customer account link")
	if err != nil {
		return nil, err
	}

	var links []models.CustomerAccount
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}
		if len(row) < 2 {
			continue // Skip incomplete rows
		}
		links = append(links, models.CustomerAccount{
			CustomerNumber: row[0],
			AccountNumber:  row[1],
		})
	}
	return links, nil
}
