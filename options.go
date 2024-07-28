package gulter

import (
	"fmt"
	"strings"
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

// ChainValidators returns a validator that accepts multiple validating criteriacriteria
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
