package gulter

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func ChainHandlers(handlers ...http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

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

	defaultFileUploadMaxSize = 1024 * 1024 * 5
)

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
