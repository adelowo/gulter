# Gulter

Gulter is a Go HTTP middleware designed to simplify the process of uploading files for your web apps. It follows the standard
`http.Handler` and `http.HandlerFunc` interfaces so you can always use with any of framework or the standard library router.

> Name and idea was gotten from the insanely popular multer package in Node.JS that does the same.

### Installation

```go

go get -u -v github.com/adelowo/gulter

```

### Usage

#### Standard HTTP router

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/adelowo/gulter"
	"github.com/adelowo/gulter/storage"
)

func main() {
	s3Store, err := storage.NewS3FromEnvironment(storage.S3Options{
		Bucket: "std-router",
	})
	if err != nil {
		panic(err.Error())
	}

	// diskStore,err := storage.NewDiskStorage("/Users/lanreadelowo/gulter-uploads/")

	handler := gulter.New(
		gulter.WithMaxFileSize(10<<20),
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

```


#### Chi router and other compatible HTTP handlers


```go

	s3Store, err := storage.NewS3FromEnvironment(storage.S3Options{
		Bucket: "chi-router",
	})
	if err != nil {
		panic(err.Error())
	}

	// diskStore,err := storage.NewDiskStorage("/Users/lanreadelowo/gulter-uploads/")

	handler := gulter.New(
		gulter.WithMaxFileSize(10<<20),
		gulter.WithValidationFunc(gulter.ChainValidators(gulter.MimeTypeValidator("image/jpeg", "image/png"))),
		gulter.WithStorage(s3Store),
	)

	router := chi.NewMux()

  // upload all files in the form fields called "form-field-1" and "form-field-2"
	router.With(handler.Upload("form-field-1", "form-field-2")).Post("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Uploaded file")

		f, err := gulter.FilesFromContext(r)
		if err != nil {
			fmt.Println(err)
			return
		}

		ff, err := gulter.FileFromContext(r, "form-field-1") // or form-field-2
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("%+v", ff)

		for _, v := range f {
			fmt.Printf("%+v", v)
			fmt.Println()
		}
	})

```

