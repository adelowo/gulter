package gulter

import (
	"context"
	"io"
)

type Storage interface {
	// Upload copies the reader to the backend file storage
	// The name of the file is also provided.
	Upload(context.Context, io.Reader, string) error
	io.Closer
}
