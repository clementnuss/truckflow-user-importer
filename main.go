package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	// "io"

	// "path/filepath"

	"github.com/clementnuss/truckflow-user-importer/internal"
	"github.com/clementnuss/truckflow-user-importer/internal/truckflow"
)

func main() {
	// Initialize database connection
	db, err := internal.InitDB()
	if err != nil {
		fmt.Printf("Database initialization error: %v\n", err)
		return
	}
	defer db.Close()

	// Open the CSV file
	file, err := os.Open("records.csv")
	if err != nil {
		fmt.Printf("Error opening CSV file: %v\n", err)
		return
	}
	defer file.Close()

	transactions, err := internal.ParseCSV(file)
	if err != nil {
		slog.Error("Error parsing CSV transactions header: %v\n", "error", err)
		return
	}
	code := 1

	for _, tr := range transactions {
		if tr.Status == "Cancelled" {
			continue
		}

		processed, err := isTransactionProcessed(db, generateHash(tr.Email), tr.Id)
		if err != nil {
			slog.Error("Unable to check if transaction has already been processed", "transaction", tr, "error", err)
			continue
		}

		if processed {
			continue
		}

		label := ""
		if tr.Company != "" {
			label = tr.Company
		} else {
			label = fmt.Sprintf("%s %s", tr.FirstName, tr.LastName)
		}

		code += 1
		tiers := truckflow.Tiers{
			Type:      "Client",
			Label:     label,
			Active:    true,
			Address:   tr.StreeAndNo,
			ZIPCode:   tr.ZIPCode,
			City:      tr.City,
			Telephone: tr.Telephone,
			Code:      fmt.Sprintf("%05d", code),
		}
		truckflowImport := truckflow.Import{
			Version: "1.50",
			Items:   []truckflow.Tiers{tiers},
		}

		os.MkdirAll("output", os.ModePerm)
		jsonData, err := json.Marshal(truckflowImport)
		if err != nil {
			slog.Error("Error marshaling JSON", "transaction", tr, "error", err)
			continue
		}

		path := filepath.Join("output", fmt.Sprintf("form_import_%s.json", tiers.Code))
		err = os.WriteFile(path, jsonData, 0644)
		if err != nil {
			slog.Error("unable to write tiers export file", "path", path, "error", err)
			continue
		}

	}

	// 	// Record processed transaction
	// 	err = recordProcessedTransaction(db, clientHash, transactionID)
	// 	if err != nil {
	// 		fmt.Printf("Error recording processed transaction: %v\n", err)
	// 		continue
	// 	}
	// }

	// Generate JSON files for each client
	// for hash, client := range clients {
	// 	filename := filepath.Join("output", fmt.Sprintf("client_%s.json", hash))
	//
	// 	// Create output directory if it doesn't exist
	// 	os.MkdirAll("output", os.ModePerm)
	//
	// 	// Create and write JSON file
	// 	jsonData, err := json.MarshalIndent(client, "", "    ")
	// 	if err != nil {
	// 		fmt.Printf("Error marshaling JSON for client %s: %v\n", client.Email, err)
	// 		continue
	// 	}
	//
	// 	err = os.WriteFile(filename, jsonData, 0644)
	// 	if err != nil {
	// 		fmt.Printf("Error writing JSON file for client %s: %v\n", client.Email, err)
	// 		continue
	// 	}
	// }
}

func generateHash(email string) string {
	hasher := sha256.New()
	hasher.Write([]byte(email))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func isTransactionProcessed(db *sql.DB, clientHash, transactionID string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM processed_records WHERE client_hash = ? AND transaction_id = ?)",
		clientHash, transactionID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func recordProcessedTransaction(db *sql.DB, clientHash, transactionID string) error {
	_, err := db.Exec(
		"INSERT INTO processed_records (client_hash, transaction_id) VALUES (?, ?)",
		clientHash, transactionID,
	)
	return err
}
