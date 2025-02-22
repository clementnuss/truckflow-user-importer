package webhook

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/clementnuss/truckflow-user-importer/internal/database"
	"github.com/clementnuss/truckflow-user-importer/internal/payrexx"
	"github.com/clementnuss/truckflow-user-importer/internal/truckflow"
	"github.com/minio/minio-go/v7"
)

var mx sync.Mutex = sync.Mutex{}

func WebhookHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	mx.Lock()
	defer mx.Unlock()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	formData := struct {
		Transaction payrexx.Transaction `json:"transaction"`
	}{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &formData); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	transaction := formData.Transaction
	transaction.SanitizeFields()
	if transaction.Status != "confirmed" {
		slog.Info("skipping uncompleted transaction", "status", transaction.Status)
		_, _ = w.Write([]byte("Ignoring uncompleted transaction"))
		return
	}

	clientHash := database.GenerateHash(transaction.Contact.Email)

	processed, err := database.IsTransactionProcessed(db, clientHash, transaction.Uuid)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if processed {
		slog.Info("skipping already processed transaction.", "transaction", transaction.Uuid)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Transaction already processed"))
		return
	}

	clientCounter, err := database.RetrieveCounter(db, "client")
	if err != nil {
		slog.Error("unable to retrieve the client counter from the DB", "error", err)
		http.Error(w, "Database error - couldn't retrieve client counter", http.StatusInternalServerError)
		return
	}

	// tier creation
	clientCounter += 1
	tiers := truckflow.Tiers{
		Type:         "Fournisseur",
		Active:       true,
		Address:      transaction.Contact.StreetAndNo,
		ZIPCode:      transaction.Contact.ZIPCode,
		City:         transaction.Contact.City,
		Telephone:    transaction.Contact.Telephone,
		Email:        transaction.Contact.Email,
		Code:         fmt.Sprintf("%05d", clientCounter),
		ProductCodes: "Dechets verts",
	}

	if transaction.Contact.ClientType == payrexx.Company {
		tiers.Label = transaction.Contact.Company
		tiers.ContactPerson = transaction.Contact.FirstName + " " + transaction.Contact.LastName
	} else {
		tiers.Label = transaction.Contact.FirstName + " " + transaction.Contact.LastName
	}
	truckflowImport := truckflow.TiersImport{
		Version: "1.50",
		Items:   []truckflow.Tiers{tiers},
	}

	jsonData, err := json.Marshal(truckflowImport)
	if err != nil {
		slog.Error("Error marshaling JSON", "transaction", transaction.Uuid, "error", err)
		http.Error(w, "Error marshaling JSON for tier", http.StatusInternalServerError)
		return
	}

	path := filepath.Join("importer/", fmt.Sprintf("tiers_import_%s.json", tiers.Code))
	_, err = s3.PutObject(
		context.Background(),
		os.Getenv("S3_BUCKET"),
		path,
		bytes.NewReader(jsonData),
		int64(len(jsonData)),
		minio.PutObjectOptions{},
	)
	if err != nil {
		slog.Error("unable to put json file on s3 bucket", "object", path, "error", err)
		http.Error(w, fmt.Sprintf("unable to write tiers json. error: %v", err), http.StatusInternalServerError)
		return
	}
	_ = database.SetCounter(db, "client", clientCounter)

	// pass creation
	passImport := truckflow.PassImport{
		Version: "1.50",
		Culture: "fr",
	}
	passCounter, err := database.RetrieveCounter(db, "pass")
	if err != nil {
		slog.Error("unable to retrieve the pass counter from the DB", "error", err)
		http.Error(w, "Database error - couldn't retrieve pass counter", http.StatusInternalServerError)
		return
	}
	for _, pl := range transaction.Plates {
		passCounter += 1
		pa := truckflow.NewPass()
		pa.Plate = pl
		pa.ParkCode = fmt.Sprintf("NEW%05d", passCounter)
		pa.Label = pa.ParkCode
		pa.TiersCode = tiers.Code

		switch transaction.Contact.ClientType {
		case payrexx.Company:
			pa.CompanyCode = "entreprises"
		case payrexx.Individual:
			pa.CompanyCode = "particuliers"
		default:
			slog.Error("unknown client type. assigning individual type", "transactionId", transaction.Uuid)
			pa.CompanyCode = "particuliers"
		}
		passImport.Items = append(passImport.Items, *pa)
	}
	jsonData, err = json.Marshal(passImport)
	if err != nil {
		slog.Error("Error marshaling JSON", "transaction", transaction.Uuid, "error", err)
		http.Error(w, "Error marshaling JSON for tier", http.StatusInternalServerError)
		return
	}

	path = filepath.Join("importer/", fmt.Sprintf("pass_import_%s.json", tiers.Code))
	_, err = s3.PutObject(
		context.Background(),
		os.Getenv("S3_BUCKET"),
		path,
		bytes.NewReader(jsonData),
		int64(len(jsonData)),
		minio.PutObjectOptions{},
	)
	if err != nil {
		slog.Error("unable to put json file on s3 bucket", "object", path, "error", err)
		http.Error(w, fmt.Sprintf("unable to write pass json. error: %v", err), http.StatusInternalServerError)
		return
	}
	_ = database.SetCounter(db, "pass", passCounter)

	err = database.RecordProcessedTransaction(db, clientHash, transaction.Uuid)
	if err != nil {
		slog.Error("unable to record processed transaction.", "transaction", transaction.Uuid, "error", err, "code", tiers.Code, "label", tiers.Label)
	}

	slog.Info("successfully imported a new tier", "tiers", transaction.Uuid, "code", tiers.Code, "label", tiers.Label)
}
