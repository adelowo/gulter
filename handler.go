package gulter

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/sync/errgroup"
)

type File struct {
	// FieldName denotes the field from the multipart form
	FieldName string `json:"field_name,omitempty"`

	// The name of the file from the client side
	OriginalName string `json:"original_name,omitempty"`
	// UploadedFileName denotes the name of the file when it was ultimately
	// uploaded to the storage layer. The distinction is important because of
	// potential changes to the file nmae that may be done
	UploadedFileName string `json:"uploaded_file_name,omitempty"`
	// FolderDestination is the folder that holds the uploaded file
	FolderDestination string `json:"folder_destination,omitempty"`

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
// it is ultimately uplaoded and stored. This is neccessarily if
// you have to adhere to specific formats as an example
type NameGeneratorFunc func(s string) string

type Gulter struct {
	storage              Storage
	destination          string
	maxSize              int64
	formKeys             []string
	validationFunc       ValidationFunc
	nameFuncGenerator    NameGeneratorFunc
	errorResponseHandler ErrResponseHandler
}

func New(opts ...Option) *Gulter {
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

	return handler
}

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
				// still need this for users pre go 1.22
				func(key string) {
					wg.Go(func() error {
						f, header, err := r.FormFile(key)
						if err != nil {
							return err
						}

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

						uploadedFiles[key] = fileData
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

	return contentType, nil
}
