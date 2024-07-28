# Gulter

[![Go Reference](https://pkg.go.dev/badge/github.com/adelowo/gulter.svg)](https://pkg.go.dev/github.com/adelowo/gulter)
[![Go Report Card](https://goreportcard.com/badge/github.com/adelowo/gulter)](https://goreportcard.com/report/github.com/adelowo/gulter)

Gulter is a Go HTTP middleware designed to simplify the process of uploading files
for your web apps. It follows the standard
`http.Handler` and `http.HandlerFunc` interfaces so you can
always use with any of framework or the standard library router.

Multiple files per form field are already supported

> Name and idea was gotten from the insanely popular multer package
> in NodeJS that does the same.

## Installation

```sh

go get -u -v github.com/adelowo/gulter

```

## Usage

Assuming you have a HTML form like this:

```html
<form action="/" method="post" enctype="multipart/form-data">
  <input type="file" name="form-field-1" />
  <input type="file" name="form-field-2" />
</form>
```

To create a new Gulter instance, you can do something like this:

```go
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
  gulter.WithStorage(s3Store),
 )
```

The `handler` is really just a HTTP middleware with the following signature
`Upload(keys ...string) func(next http.Handler) http.Handler`. `keys` here
are the input names from the HTML form, so you can chain this into almost any HTTP
router.

### Standard HTTP router

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

 handler, err := gulter.New(
  gulter.WithMaxFileSize(10<<20),
  gulter.WithStorage(s3Store),
 )

 mux := http.NewServeMux()

 // upload all files with the "name" and "lanre" fields on this route
 mux.Handle("/", handler.Upload("name", "lanre")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  fmt.Println("Uploaded file")

  // return all uploaded files
  f, err := gulter.FilesFromContext(r)
  if err != nil {
   fmt.Println(err)
   return
  }

  // return uploaded files with the form field "lanre"
  ff, err := gulter.FilesFromContextWithKey(r, "lanre")
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

### Chi router and other compatible HTTP handlers

```go

 s3Store, err := storage.NewS3FromEnvironment(storage.S3Options{
  Bucket: "chi-router",
 })
 if err != nil {
  panic(err.Error())
 }

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

  ff, err := gulter.FilesFromContextWithKey(r, "form-field-1") // or form-field-2
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

## API

While this middleware automatically uploads your files, sometimes you need
details about the uploaded file to show to the user, this could be making up the
image url or path to the image. To get that in your HTTP handler, you can use either:

- `FilesFromContextWithKey`: retrieve a named input uploaded file.
- `FilesFromContext`: retrieve all uploaded files

Gulter also ships with two storage implementations at the moment:

- `S3Store` : supports S3 or any compatible service like Minio,R2 & others
- `DiskStore`: uses a local filesystem backed store to upload files
- `CloudinaryStore`: uploads file to cloudinary

## Ignoring non existent keys in the multipart Request

Sometimes, the keys you have configured the middleware might get dropped from the
frontend for some reason, ideally the middleware fails if it cannot find a
configured key in the request. To disable this behavior and ignore the missing
key, you can make use of the `WithIgnoreNonExistentKey(true)` option to prevent the
middleware from causing an error when such keys do not exists

## Customizing the error response

Since Gulter is a middleware that runs, it returns an error to the client if found,
this might not match your existing structure, so to configure the response, use the
`WithErrorResponseHandler`. The default is shown below and can be used as a template
to define yours.

```go

 errHandler ErrResponseHandler = func(err error) http.HandlerFunc {
  return func(w http.ResponseWriter, _ *http.Request) {
   w.Header().Set("Content-Type", "application/json")
   w.WriteHeader(http.StatusInternalServerError)
   fmt.Fprintf(w, `{"message" : "could not upload file", "error" : %s}`, err.Error())
  }
 }
```

## Writing your custom validator logic

Sometimes, you could have some custom logic to validate uploads, in this example
below, we limit the size of the upload based on the mimeypes of the uploaded files

```go

var customValidator gulter.ValidationFunc = func(f gulter.File) error {
 switch f.MimeType {
 case "image/png":
  if f.Size > 4096 {
   return errors.New("file size too large")
  }

  return nil

 case "application/pdf":
  if f.Size > (1024 * 10) {
   return errors.New("file size too large")
  }

  return nil
 default:
  return nil
 }
}

```
