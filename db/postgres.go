package db

import (
	"database/sql"
	"fmt"
	"log"

	"importer/config"
	"importer/models"
)

type PostgresDB struct {
	db  *sql.DB
	cfg *config.AppConfig
}

func NewPostgresDB(cfg *config.AppConfig) (*PostgresDB, error) {
	db, err := sql.Open("postgres", cfg.DB.ConnectionString())
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &PostgresDB{
		db:  db,
		cfg: cfg,
	}, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// CustomerRepository implementation
func (p *PostgresDB) InsertCustomers(customers []models.Customer) (map[string]int, error) {
	customerIDs := make(map[string]int)
	tx, err := p.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}

	stmt, err := tx.Prepare(`
        INSERT INTO customers (client_id, customer_number, customer_name, address, name, email)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (customer_number) DO UPDATE SET
            client_id = EXCLUDED.client_id,
            customer_name = EXCLUDED.customer_name,
            address = EXCLUDED.address,
            name = EXCLUDED.name,
            email = EXCLUDED.email,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id, customer_number`)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for i, customer := range customers {
		var id int
		err = stmt.QueryRow(
			customer.ClientID,
			customer.CustomerNumber,
			customer.CustomerName,
			customer.Address,
			customer.Name,
			customer.Email,
		).Scan(&id, &customer.CustomerNumber)

		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to insert customer %s: %v", customer.CustomerNumber, err)
		}

		customerIDs[customer.CustomerNumber] = id

		if (i+1)%p.cfg.BatchSize == 0 {
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to commit batch: %v", err)
			}
			tx, err = p.db.Begin()
			if err != nil {
				return nil, fmt.Errorf("failed to begin new transaction: %v", err)
			}
			stmt, err = tx.Prepare(`
                INSERT INTO customers (client_id, customer_number, customer_name, address, name, email)
                VALUES ($1, $2, $3, $4, $5, $6)
                ON CONFLICT (customer_number) DO UPDATE SET
                    client_id = EXCLUDED.client_id,
                    customer_name = EXCLUDED.customer_name,
                    address = EXCLUDED.address,
                    name = EXCLUDED.name,
                    email = EXCLUDED.email,
                    updated_at = CURRENT_TIMESTAMP
                RETURNING id, customer_number`)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to prepare statement: %v", err)
			}
			log.Printf("Processed %d customers", i+1)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit final batch: %v", err)
	}

	return customerIDs, nil
}

// AccountRepository implementation
func (p *PostgresDB) InsertAccounts(accounts []models.Account) (map[string]int, error) {
	accountIDs := make(map[string]int)
	tx, err := p.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}

	stmt, err := tx.Prepare(`
        INSERT INTO accounts (account_number, account_name)
        VALUES ($1, $2)
        ON CONFLICT (account_number) DO UPDATE SET
            account_name = EXCLUDED.account_name,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id, account_number`)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for i, account := range accounts {
		var id int
		err = stmt.QueryRow(
			account.AccountNumber,
			account.AccountName,
		).Scan(&id, &account.AccountNumber)

		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to insert account %s: %v", account.AccountNumber, err)
		}

		accountIDs[account.AccountNumber] = id

		if (i+1)%p.cfg.BatchSize == 0 {
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to commit batch: %v", err)
			}
			tx, err = p.db.Begin()
			if err != nil {
				return nil, fmt.Errorf("failed to begin new transaction: %v", err)
			}
			stmt, err = tx.Prepare(`
                INSERT INTO accounts (account_number, account_name)
                VALUES ($1, $2)
                ON CONFLICT (account_number) DO UPDATE SET
                    account_name = EXCLUDED.account_name,
                    updated_at = CURRENT_TIMESTAMP
                RETURNING id, account_number`)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to prepare statement: %v", err)
			}
			log.Printf("Processed %d accounts", i+1)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit final batch: %v", err)
	}

	return accountIDs, nil
}

// CustomerAccountRepository implementation
func (p *PostgresDB) InsertCustomerAccounts(links []models.CustomerAccount, customerIDs, accountIDs map[string]int) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	stmt, err := tx.Prepare(`
        INSERT INTO customer_accounts (customer_id, account_id)
        VALUES ($1, $2)
        ON CONFLICT (customer_id, account_id) DO NOTHING`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for i, link := range links {
		customerID, ok := customerIDs[link.CustomerNumber]
		if !ok {
			log.Printf("Warning: Customer number %s not found", link.CustomerNumber)
			continue
		}

		accountID, ok := accountIDs[link.AccountNumber]
		if !ok {
			log.Printf("Warning: Account number %s not found", link.AccountNumber)
			continue
		}

		_, err = stmt.Exec(customerID, accountID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert customer-account link %s-%s: %v",
				link.CustomerNumber, link.AccountNumber, err)
		}

		if (i+1)%p.cfg.BatchSize == 0 {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit batch: %v", err)
			}
			tx, err = p.db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin new transaction: %v", err)
			}
			stmt, err = tx.Prepare(`
                INSERT INTO customer_accounts (customer_id, account_id)
                VALUES ($1, $2)
                ON CONFLICT (customer_id, account_id) DO NOTHING`)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to prepare statement: %v", err)
			}
			log.Printf("Processed %d customer-account links", i+1)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit final batch: %v", err)
	}

	return nil
}
