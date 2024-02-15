package main

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/adelowo/gulter"
	"github.com/adelowo/gulter/storage"
)

func main() {
	handler := gulter.New(
		gulter.WithDestination("/Users/lanreadelowo/yikes"),
		gulter.WithMaxFileSize(10<<20),
		gulter.WithNameFuncGenerator(func(s string) string {
			return "gulter-" + s
		}),
		gulter.WithValidationFunc(func(f multipart.File) error {
			return nil
		}),
		gulter.WithStorage(&storage.Disk{}))

	mux := http.NewServeMux()

	mux.Handle("/", handler.Upload("name")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Uploaded file")
	})))

	http.ListenAndServe(":3300", mux)
}
