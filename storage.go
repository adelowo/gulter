package gulter

import (
	"context"
	"io"
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

type Storage interface {
	// Upload copies the reader to the backend file storage
	// The name of the file is also provided.
	Upload(context.Context, io.Reader, *UploadFileOptions) (*UploadedFileMetadata, error)
	io.Closer
}
