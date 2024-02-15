package gulter

import (
	"fmt"
	"mime/multipart"
	"net/http"
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

type Gulter struct {
	storage           Storage
	destination       string
	maxSize           int64
	formKeys          []string
	validationFunc    func(f multipart.File) error
	nameFuncGenerator func(s string) string
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

				if err := h.storage.Upload(r.Context(), f,
					h.nameFuncGenerator(header.Filename)); err != nil {

					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(w, `{"error" : "could not upload file with key(%s).. %v"}`, key, err)
					return
				}

				f.Close()
			}

			next.ServeHTTP(w, r)
		})
	}
}
