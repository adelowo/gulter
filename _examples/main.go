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

	mux.Handle("/", handler.Upload("name", "lanre")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Uploaded file")

		f, err := gulter.FilesFromContext(r)
		if err != nil {
			fmt.Println(err)
			return
		}

		ff, err := gulter.FileFromContext(r, "lanre")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("%+v", ff)

		for _, v := range f {
			fmt.Printf("%+v", v)
			fmt.Println()
		}
	})))

	http.ListenAndServe(":3300", mux)
}
