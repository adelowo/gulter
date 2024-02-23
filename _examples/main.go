package main

import (
	"fmt"
	"net/http"

	"github.com/adelowo/gulter"
	"github.com/adelowo/gulter/storage"
)

func main() {
	s3Store, err := storage.NewS3FromEnvironment(storage.S3Options{
		Bucket: "fotion",
	})
	if err != nil {
		panic(err.Error())
	}

	// diskStore,err := storage.NewDiskStorage("/Users/lanreadelowo/gulter-uploads/")

	handler := gulter.New(
		gulter.WithMaxFileSize(10<<20),
		// gulter.WithValidationFunc(gulter.ChainValidators(gulter.MimeTypeValidator("image/jpeg"))),
		gulter.WithStorage(s3Store),
	)

	mux := http.NewServeMux()

	// upload all files with the "name" and "lanre" fields on this route
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
