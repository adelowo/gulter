package main

import (
	"fmt"
	"net/http"

	"github.com/adelowo/gulter"
	"github.com/adelowo/gulter/storage"
	"github.com/google/uuid"
)

func main() {

	disk, _ := storage.NewDiskStorage("/Users/lanreadelowo/yikes/")

	// do not ignore :))
	handler, _ := gulter.New(
		gulter.WithMaxFileSize(10<<20),
		gulter.WithValidationFunc(
			gulter.ChainValidators(gulter.MimeTypeValidator("image/jpeg", "image/png"),
				func(f gulter.File) error {
					// Your own custom validation function on the file here
					// Else you can really just drop the ChainValidators and use only the MimeTypeValidator or just
					// one custom validator alone
					return nil
				})),
		gulter.WithNameFuncGenerator(func(s string) string {
			return uuid.NewString()
		}),
		gulter.WithStorage(disk),
	)

	mux := http.NewServeMux()

	// upload all files with the "name" and "lanre" fields on this route
	mux.Handle("/", handler.Upload("bucket_name", "lanre")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Uploaded file")

		f, err := gulter.FilesFromContext(r)
		if err != nil {
			fmt.Println(err)
			return
		}

		ff, err := gulter.FilesFromContextWithKey(r, "lanre")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("%+v", ff)

		for _, v := range f {
			fmt.Printf("%+v\n", v)
			fmt.Println()
		}
	})))

	http.ListenAndServe(":3300", mux)
}
