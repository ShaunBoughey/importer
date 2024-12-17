package models

import "fmt"

// API Request Models
type CustomerRequest struct {
	ClientID       string `json:"client_id,omitempty"`
	CustomerNumber string `json:"customer_number"`   // Required
	CustomerName   string `json:"customer_name"`     // Required
	Address        string `json:"address,omitempty"` // Optional
	Name           string `json:"name,omitempty"`    // Optional
	Email          string `json:"email,omitempty"`   // Optional
}

type AccountRequest struct {
	AccountNumber string `json:"account_number"` // Required
	AccountName   string `json:"account_name"`   // Required
}

type CustomerAccountLinkRequest struct {
	CustomerID int `json:"customer_id"` // Required
	AccountID  int `json:"account_id"`  // Required
}

// Validation Error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Conversion functions
func ToCustomerRequest(c Customer) CustomerRequest {
	return CustomerRequest{
		ClientID:       emptyToNil(c.ClientID),
		CustomerNumber: c.CustomerNumber,
		CustomerName:   c.CustomerName,
		Address:        emptyToNil(c.Address),
		Name:           emptyToNil(c.Name),
		Email:          emptyToNil(c.Email),
	}
}

func ToAccountRequest(a Account) AccountRequest {
	return AccountRequest{
		AccountNumber: a.AccountNumber,
		AccountName:   a.AccountName,
	}
}

func ToLinkRequest(customerID, accountID int) CustomerAccountLinkRequest {
	return CustomerAccountLinkRequest{
		CustomerID: customerID,
		AccountID:  accountID,
	}
}

// Helper functions
func emptyToNil(s string) string {
	if s == "" {
		return ""
	}
	return s
}
