package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/adelowo/gulter"
	"github.com/adelowo/gulter/storage"
	"github.com/google/uuid"
)

func main() {
	storageAccount := mustGetenv("AZURE_STORAGE_ACCOUNT")
	containerName := mustGetenv("AZURE_STORAGE_CONTAINER")
	serviceURL := os.Getenv("AZURE_BLOB_SERVICE_URL")
	if serviceURL == "" {
		serviceURL = fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccount)
	}

	azureBlobStore, err := storage.NewAzureBlobFromCredential(storage.AzureBlobOptions{
		ContainerName:  containerName,
		ServiceUrl:     serviceURL,
		StorageAccount: storageAccount,
	})
	if err != nil {
		log.Fatal(err)
	}

	handler, err := gulter.New(
		gulter.WithMaxFileSize(10<<20),
		gulter.WithValidationFunc(gulter.MimeTypeValidator("image/jpeg", "image/png")),
		gulter.WithNameFuncGenerator(func(s string) string {
			return uuid.NewString() + filepath.Ext(s)
		}),
		gulter.WithStorage(azureBlobStore),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.Handle("/azure", handler.Upload("file")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files, err := gulter.FilesFromContextWithKey(r, "file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, file := range files {
			path, err := azureBlobStore.Path(context.Background(), gulter.PathOptions{
				Key: file.StorageKey,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "%s\n", path)
		}
	})))

	log.Println("listening on :3400")
	log.Fatal(http.ListenAndServe(":3400", mux))
}

func mustGetenv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is required", key)
	}

	return value
}
