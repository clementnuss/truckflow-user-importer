package internal

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)

func InitDB() (*sql.DB, error) {
	// Update these values according to your MariaDB configuration
  dsn := "root:dev@tcp(localhost:3306)/truckflow_importer"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	// Create the processed_records table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS processed_records (
            id INT AUTO_INCREMENT PRIMARY KEY,
            client_hash VARCHAR(32) NOT NULL,
            transaction_id VARCHAR(255) NOT NULL,
            processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE KEY unique_transaction (client_hash, transaction_id)
        )
    `)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %v", err)
	}

	return db, nil
}
