package gulter

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
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

	MimeType string `json:"mime_type,omitempty"`

	Size int64 `json:"size,omitempty"`
}

func Chain(handlers ...http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

// ValidationFunc is a type that can be used to dynamically validate a file
type ValidationFunc func(f multipart.File) error

// NameGeneratorFunc allows you alter the name of the file before
// it is ultimately uplaoded and stored. This is neccessarily if
// you have to adhere to specific formats as an example
type NameGeneratorFunc func(s string) string

var (
	// allows all file pass through
	defaultValidationFunc ValidationFunc = func(f multipart.File) error {
		return nil
	}

	// defaultNameGeneratorFunc uses the gulter-158888-originalname to
	// upload files
	defaultNameGeneratorFunc NameGeneratorFunc = func(s string) string {
		return fmt.Sprintf("gulter-%d-%s", time.Now().Unix(), s)
	}

	defaultFileUploadMaxSize = 1024 * 1024 * 5
)

type Gulter struct {
	storage           Storage
	destination       string
	maxSize           int64
	formKeys          []string
	validationFunc    ValidationFunc
	nameFuncGenerator NameGeneratorFunc
}

func New(opts ...Option) *Gulter {
	handler := &Gulter{}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

func (h *Gulter) Upload(keys ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

			_ = r.ParseMultipartForm(h.maxSize)

			for _, key := range keys {
				f, header, err := r.FormFile(key)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(w, `{"error" : "could not fetch form field (%s)...%v"}`, key, err)
					return
				}

				if err := h.validationFunc(f); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(w, `{"error" : "could not validate file with key(%s).. %v"}`, key, err)
					return
				}

				uploadedFileName := h.nameFuncGenerator(header.Filename)

				var mimeType string

				metadata, err := h.storage.Upload(r.Context(), f, &UploadFileOptions{
					FileName: uploadedFileName,
				})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(w, `{"error" : "could not upload file with key(%s).. %v"}`, key, err)
					return
				}

				uploadedFile := File{
					FieldName:         key,
					OriginalName:      header.Filename,
					UploadedFileName:  uploadedFileName,
					FolderDestination: metadata.FolderDestination,
					MimeType:          mimeType,
					Size:              header.Size,
				}

				r = r.WithContext(writeFileToContext(r.Context(), uploadedFile, key))

				f.Close()
			}

			next.ServeHTTP(w, r)
		})
	}
}
