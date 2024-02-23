package gulter

import (
	"context"
	"io"
)

type UploadFileOptions struct {
	FileName string
	Metadata map[string]string
}

type UploadedFileMetadata struct {
	FolderDestination string `json:"folder_destination"`
	Size              int64  `json:"size"`
}

type Storage interface {
	// Upload copies the reader to the backend file storage
	// The name of the file is also provided.
	Upload(context.Context, io.Reader, *UploadFileOptions) (*UploadedFileMetadata, error)
	io.Closer
}
