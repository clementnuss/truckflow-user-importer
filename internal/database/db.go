package database

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB() (*sql.DB, error) {
	host := os.Getenv("MARIADB_HOST")
	user := os.Getenv("MARIADB_USER")
	password := os.Getenv("MARIADB_PASSWORD")
	database := os.Getenv("MARIADB_DATABASE")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS processed_records (
            id INT AUTO_INCREMENT PRIMARY KEY,
            transaction_id VARCHAR(32) NOT NULL,
            client_hash VARCHAR(32) NOT NULL,
            processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE KEY unique_transaction (client_hash, transaction_id)
        )
    `)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS counters (
            name varchar(32) NOT NULL PRIMARY KEY,
            value int(11) NOT NULL
        )
    `)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %v", err)
	}
	return db, nil
}

func RetrieveCounter(db *sql.DB, counter string) (int, error) {
	var v int
	err := db.QueryRow("SELECT value FROM counters WHERE name = ?", counter).Scan(&v)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, SetCounter(db, counter, 0)
	} else if err != nil {
		return -1, err
	}

	return v, nil
}

func SetCounter(db *sql.DB, counter string, value int) error {
	_, err := db.Exec("REPLACE INTO counters (name, value) VALUES (?, ?);", counter, value)

	return err
}

func IsTransactionProcessed(db *sql.DB, clientHash, transactionID string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM processed_records WHERE client_hash = ? AND transaction_id = ?)",
		clientHash, transactionID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func RecordProcessedTransaction(db *sql.DB, clientHash, transactionID string) error {
	_, err := db.Exec(
		"INSERT INTO processed_records (client_hash, transaction_id) VALUES (?, ?)",
		clientHash, transactionID,
	)
	return err
}

func GenerateHash(email string) string {
	hasher := sha256.New()
	hasher.Write([]byte(email))
	return fmt.Sprintf("%x", hasher.Sum(nil)[:8])
}
