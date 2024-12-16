package models

type Customer struct {
	ClientID       string
	CustomerNumber string
	CustomerName   string
	Address        string
	Name           string
	Email          string
}

type Account struct {
	AccountNumber string
	AccountName   string
}

type CustomerAccount struct {
	CustomerNumber string
	AccountNumber  string
}

// Repository interfaces for database operations
type CustomerRepository interface {
	InsertCustomers(customers []Customer) (map[string]int, error)
	Close() error
}

type AccountRepository interface {
	InsertAccounts(accounts []Account) (map[string]int, error)
}

type CustomerAccountRepository interface {
	InsertCustomerAccounts(links []CustomerAccount, customerIDs, accountIDs map[string]int) error
}
