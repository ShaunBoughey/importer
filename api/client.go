// api/client.go
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"importer/config"
	"importer/models"

	"golang.org/x/time/rate"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	limiter    *rate.Limiter
}

var _ models.CustomerRepository = (*Client)(nil)

func (c *Client) Close() error {
	// Clean up any resources if needed
	c.httpClient.CloseIdleConnections()
	return nil
}

func NewClient(cfg *config.AppConfig) *Client {
	return &Client{
		baseURL: cfg.API.BaseURL,
		apiKey:  cfg.API.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		limiter: rate.NewLimiter(rate.Limit(cfg.API.RateLimit), cfg.API.RateLimit),
	}
}

func (c *Client) InsertCustomers(customers []models.Customer) (map[string]int, error) {
	customerIDs := make(map[string]int)

	for i, customer := range customers {
		// Wait for rate limiter
		err := c.limiter.Wait(context.Background())
		if err != nil {
			return nil, fmt.Errorf("rate limiter error: %v", err)
		}

		// Convert to API request format
		requestBody := models.ToCustomerRequest(customer)

		// Make API request
		payload, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("error marshaling customer: %v", err)
		}

		// Log the actual payload being sent (useful for debugging)
		log.Printf("Sending customer payload: %s", string(payload))

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/customers", c.baseURL), bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error making request: %v", err)
		}
		defer resp.Body.Close()

		// Read the response body for error reporting
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("API returned status %d for customer %s: %s",
				resp.StatusCode, customer.CustomerNumber, string(body))
		}

		// Parse response
		var result struct {
			ID int `json:"id"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error decoding response: %v, body: %s", err, string(body))
		}

		customerIDs[customer.CustomerNumber] = result.ID

		if (i+1)%100 == 0 {
			log.Printf("Processed %d/%d customers", i+1, len(customers))
		}
	}

	return customerIDs, nil
}

func (c *Client) InsertAccounts(accounts []models.Account) (map[string]int, error) {
	accountIDs := make(map[string]int)

	for i, account := range accounts {
		err := c.limiter.Wait(context.Background())
		if err != nil {
			return nil, fmt.Errorf("rate limiter error: %v", err)
		}

		requestBody := models.ToAccountRequest(account)

		payload, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("error marshaling account: %v", err)
		}

		log.Printf("Sending account payload: %s", string(payload))

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/accounts", c.baseURL), bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error making request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("API returned status %d for account %s: %s",
				resp.StatusCode, account.AccountNumber, string(body))
		}

		var result struct {
			ID int `json:"id"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error decoding response: %v, body: %s", err, string(body))
		}

		accountIDs[account.AccountNumber] = result.ID

		if (i+1)%100 == 0 {
			log.Printf("Processed %d/%d accounts", i+1, len(accounts))
		}
	}

	return accountIDs, nil
}

func (c *Client) InsertCustomerAccounts(links []models.CustomerAccount, customerIDs, accountIDs map[string]int) error {
	for i, link := range links {
		err := c.limiter.Wait(context.Background())
		if err != nil {
			return fmt.Errorf("rate limiter error: %v", err)
		}

		customerID, ok := customerIDs[link.CustomerNumber]
		if !ok {
			log.Printf("Warning: Customer %s not found, skipping link", link.CustomerNumber)
			continue
		}

		accountID, ok := accountIDs[link.AccountNumber]
		if !ok {
			log.Printf("Warning: Account %s not found, skipping link", link.AccountNumber)
			continue
		}

		requestBody := models.ToLinkRequest(customerID, accountID)

		payload, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("error marshaling link: %v", err)
		}

		log.Printf("Sending link payload: %s", string(payload))

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/customer-accounts", c.baseURL), bytes.NewBuffer(payload))
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error making request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("API returned status %d for link %s-%s: %s",
				resp.StatusCode, link.CustomerNumber, link.AccountNumber, string(body))
		}

		if (i+1)%100 == 0 {
			log.Printf("Processed %d/%d links", i+1, len(links))
		}
	}

	return nil
}
