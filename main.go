package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clementnuss/truckflow-user-importer/internal/database"
	"github.com/clementnuss/truckflow-user-importer/internal/webhook"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	_ "github.com/joho/godotenv/autoload"
)

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	defer stop()

	var err error
	db, err := database.InitDB()
	if err != nil {
		slog.Error("database initialization error", "error", err)
		return
	}
	defer db.Close()

  slog.Info("database successfully initialized")

	endpoint := os.Getenv("S3_ENDPOINT")
	accessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	bucket := os.Getenv("S3_BUCKET")

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		slog.Error("database initialization error", "error", err)
		return
	}

  testData := []byte(fmt.Sprintf("test string %v", time.Now()))
	_, err = minioClient.PutObject(ctx, bucket, "importer/test", bytes.NewReader(testData), int64(len(testData)), minio.PutObjectOptions{})
  if err != nil {
    slog.Error("unable to create test file on S3 endpoint", "error", err)
    return
  }
	err = minioClient.RemoveObject(ctx, bucket, "importer/test", minio.RemoveObjectOptions{})
  if err != nil {
    slog.Error("unable to delete test file on S3 endpoint", "error", err)
    return
  }

	slog.Info("minio s3 client started")

	port := ":9000"
	server := http.Server{
		Addr: port,
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		webhook.WebhookHandler(w, r, db, minioClient)
	})

	slog.Info("webhook server starting", "port", port)
	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server cannot start listening", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	if err := server.Shutdown(context.Background()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("could not shutdown http server properly", "error", err)
		os.Exit(1)
	}

	slog.Info("graceful shutdown completed")
}
