package gulter

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"
)

type File struct {
	// FieldName denotes the field from the multipart form
	FieldName string `json:"field_name,omitempty"`

	// The name of the file from the client side
	OriginalName string `json:"original_name,omitempty"`
	// UploadedFileName denotes the name of the file when it was ultimately
	// uploaded to the storage layer. The distinction is important because of
	// potential changes to the file name that may be done
	UploadedFileName string `json:"uploaded_file_name,omitempty"`
	// FolderDestination is the folder that holds the uploaded file
	FolderDestination string `json:"folder_destination,omitempty"`

	// StorageKey can be used to retrieve the file from the storage backend
	StorageKey string `json:"storage_key,omitempty"`

	// MimeType of the uploaded file
	MimeType string `json:"mime_type,omitempty"`

	// Size in bytes of the uploaded file
	Size int64 `json:"size,omitempty"`
}

// ValidationFunc is a type that can be used to dynamically validate a file
type ValidationFunc func(f File) error

// ErrResponseHandler is a custom error that should be used to handle errors when
// an upload fails
type ErrResponseHandler func(error) http.HandlerFunc

// NameGeneratorFunc allows you alter the name of the file before
// it is ultimately uplaoded and stored. This is necessarily if
// you have to adhere to specific formats as an example
type NameGeneratorFunc func(s string) string

type Gulter struct {
	storage Storage
	maxSize int64

	// when you configure the middleware, you usually provide a list of
	// keys to retrieve the files from. If any of these keys do not exists,
	// the handler fails.
	// If this option is set to true, the value is just skipped instead
	ignoreNonExistentKeys bool

	validationFunc       ValidationFunc
	nameFuncGenerator    NameGeneratorFunc
	errorResponseHandler ErrResponseHandler
}

func New(opts ...Option) (*Gulter, error) {
	handler := &Gulter{}

	for _, opt := range opts {
		opt(handler)
	}

	if handler.maxSize <= 0 {
		handler.maxSize = defaultFileUploadMaxSize
	}

	if handler.validationFunc == nil {
		handler.validationFunc = defaultValidationFunc
	}

	if handler.nameFuncGenerator == nil {
		handler.nameFuncGenerator = defaultNameGeneratorFunc
	}

	if handler.errorResponseHandler == nil {
		handler.errorResponseHandler = defaultErrorResponseHandler
	}

	if handler.storage == nil {
		return nil, errors.New("you must provide a storage backend")
	}

	return handler, nil
}

// Upload is a HTTP middleware that takes in a list of form fields and the next
// HTTP handler to run after the upload prodcess is completed
func (h *Gulter) Upload(keys ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

			err := r.ParseMultipartForm(h.maxSize)
			if err != nil {
				h.errorResponseHandler(err).ServeHTTP(w, r)
				return
			}

			var wg errgroup.Group

			uploadedFiles := make(Files, len(keys))

			for _, key := range keys {
				// TODO(adelowo): remove this when we drop support for < 1.22
				func(key string) {
					wg.Go(func() error {

						fileHeaders, ok := r.MultipartForm.File[key]
						if !ok {
							if h.ignoreNonExistentKeys {
								return nil
							}

							return fmt.Errorf("files could not be found in key (%s) from http request", key)
						}

						for _, header := range fileHeaders {

							f, err := header.Open()

							defer f.Close()

							uploadedFileName := h.nameFuncGenerator(header.Filename)

							mimeType, err := fetchContentType(f)
							if err != nil {
								return fmt.Errorf("gulter: %s has invalid mimetype..%v", key, err)
							}

							fileData := File{
								FieldName:        key,
								OriginalName:     header.Filename,
								UploadedFileName: uploadedFileName,
								MimeType:         mimeType,
							}

							if err := h.validationFunc(fileData); err != nil {
								return fmt.Errorf("gulter: validation failed for (%s)...%v", key, err)
							}

							metadata, err := h.storage.Upload(r.Context(), f, &UploadFileOptions{
								FileName: uploadedFileName,
							})
							if err != nil {
								return fmt.Errorf("gulter: could not upload file to storage (%s)...%v", key, err)
							}

							fileData.Size = metadata.Size
							fileData.FolderDestination = metadata.FolderDestination
							fileData.StorageKey = metadata.Key

							uploadedFiles[key] = fileData
							return nil
						}

						return nil
					})
				}(key)
			}

			if err := wg.Wait(); err != nil {
				h.errorResponseHandler(err).ServeHTTP(w, r)
				return
			}

			r = r.WithContext(writeFilesToContext(r.Context(), uploadedFiles))

			next.ServeHTTP(w, r)
		})
	}
}

func fetchContentType(f io.ReadSeeker) (string, error) {
	buff := make([]byte, 512)

	_, err := f.Seek(0, 0)
	if err != nil {
		return "", err
	}

	bytesRead, err := f.Read(buff)
	if err != nil && err != io.EOF {
		return "", err
	}

	buff = buff[:bytesRead]

	contentType := http.DetectContentType(buff)

	_, err = f.Seek(0, 0)
	if err != nil {
		return "", err
	}

	// text/plain; charset=utf-8
	// we do not want users to have to specify such long mimetypes
	// Specifying text/plain should be enough really
	// If we have such mimetypes with the charset included, just strip
	// it out completely
	splitType := strings.Split(contentType, ";")
	if len(splitType) == 2 {
		contentType = splitType[0]
	}

	return contentType, nil
}
