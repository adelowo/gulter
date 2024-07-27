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

type Files map[string][]File

func writeFilesToContext(ctx context.Context,
	f Files,
) context.Context {
	existingFiles, ok := ctx.Value(fileKey).(Files)
	if !ok {
		existingFiles = Files{}
	}

	for _, v := range f {
		// all the files should have the same form field,
		// so safe to use any index
		existingFiles[v[0].FieldName] = append(existingFiles[v[0].FieldName], v...)
	}

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

// FilesFromContextWithKey returns  all files that have been uploaded during the request
// and sorts by the provided form field
func FilesFromContextWithKey(r *http.Request, key string) ([]File, error) {
	files, ok := r.Context().Value(fileKey).(Files)
	if !ok {
		return nil, ErrNoFilesUploaded
	}

	return files[key], nil
}
