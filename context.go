package gulter

import (
	"context"
	"net/http"
)

type contextKey string

const (
	fileKey contextKey = "files"
)

type errorMsg string

func (e errorMsg) Error() string { return string(e) }

const (
	ErrNoFilesUploaded = errorMsg("gulter: no uploadable files found in request")
)

type Files map[string]File

func writeFileToContext(ctx context.Context,
	f File, formField string,
) context.Context {
	existingFiles, ok := ctx.Value(fileKey).(Files)
	if !ok {
		existingFiles = Files{}
	}

	existingFiles[formField] = f
	return context.WithValue(ctx, fileKey, existingFiles)
}

// FilesFromContext returns all files that have been uploaded during the request
func FilesFromContext(r *http.Request) (Files, error) {
	files, ok := r.Context().Value(fileKey).(Files)
	if !ok {
		return nil, ErrNoFilesUploaded
	}

	return files, nil
}

// FileFromContext retrieves the uploaded file with
// the given formfield value. This form field is what
// was sent from the html/multipart form
func FileFromContext(r *http.Request,
	formField string,
) (File, error) {
	files, err := FilesFromContext(r)
	if err != nil {
		return File{}, err
	}

	f, ok := files[formField]
	if !ok {
		return File{}, ErrNoFilesUploaded
	}

	return f, nil
}
