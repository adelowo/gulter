package gulter

import (
	"context"
	"io"
	"time"
)

type UploadFileOptions struct {
	FileName string
	Metadata map[string]string

	// If not provided, the default bucket will be used
	// This is useful
	Bucket string
}

type UploadedFileMetadata struct {
	FolderDestination string `json:"folder_destination,omitempty"`
	Key               string `json:"key,omitempty"`
	Size              int64  `json:"size,omitempty"`
}

type PathOptions struct {
	Bucket string `json:"bucket,omitempty"`
	Key    string `json:"key,omitempty"`

	// Will only take effect if IsSecure is provided
	ExpirationTime time.Duration `json:"expiration_time,omitempty"`
	IsSecure       bool          `json:"is_secure,omitempty"`
}

type Storage interface {
	// Upload copies the reader to the backend file storage
	// The name of the file is also provided.
	Upload(context.Context, io.Reader, *UploadFileOptions) (*UploadedFileMetadata, error)
	Path(context.Context, string, PathOptions) (string, error)
	io.Closer
}
