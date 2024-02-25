package gulter

import (
	"fmt"
	"net/http"
	"strings"
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

// MimeTypeValidator makes sure we only accept a valid mimetype.
// It takes in an array of supported mimes
func MimeTypeValidator(validMimeTypes ...string) ValidationFunc {
	return func(f File) error {
		for _, mimeType := range validMimeTypes {
			if strings.EqualFold(strings.ToLower(mimeType), f.MimeType) {
				return nil
			}
		}
		return fmt.Errorf("unsupported mime type uploaded..(%s)", f.MimeType)
	}
}

// ChainValidators returns a validator that accepts multiple validating criterias
func ChainValidators(validators ...ValidationFunc) ValidationFunc {
	return func(f File) error {
		for _, validator := range validators {
			if err := validator(f); err != nil {
				return err
			}
		}

		return nil
	}
}
