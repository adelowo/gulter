package gulter

import (
	"fmt"
	"net/http"
	"time"
)

var (
	// allows all file pass through
	defaultValidationFunc ValidationFunc = func(f File) error {
		return nil
	}

	// defaultNameGeneratorFunc uses the gulter-158888-originalname to
	// upload files
	defaultNameGeneratorFunc NameGeneratorFunc = func(s string) string {
		return fmt.Sprintf("gulter-%d-%s", time.Now().Unix(), s)
	}

	defaultFileUploadMaxSize int64 = 1024 * 1024 * 5

	defaultErrorResponseHandler ErrResponseHandler = func(err error) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"message" : "could not upload file", "error" : %s}`, err.Error())
		}
	}
)

type Option func(*Gulter)

func WithStorage(store Storage) Option {
	return func(gh *Gulter) {
		gh.storage = store
	}
}

// WithMaxFileSize allows you limit the size of file uploads to accept
func WithMaxFileSize(i int64) Option {
	return func(gh *Gulter) {
		gh.maxSize = i
	}
}

func WithValidationFunc(validationFunc ValidationFunc) Option {
	return func(g *Gulter) {
		g.validationFunc = validationFunc
	}
}

// WithNameFuncGenerator allows you configure how you'd like to rename your
// uploaded files
func WithNameFuncGenerator(nameFunc NameGeneratorFunc) Option {
	return func(g *Gulter) {
		g.nameFuncGenerator = nameFunc
	}
}

func WithIgnoreNonExistentKey(ignore bool) Option {
	return func(g *Gulter) {
		g.ignoreNonExistentKeys = ignore
	}
}

func WithErrorResponseHandler(errHandler ErrResponseHandler) Option {
	return func(g *Gulter) {
		g.errorResponseHandler = errHandler
	}
}
