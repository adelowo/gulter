package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/adelowo/gulter"
)

type Disk struct {
}

func NewDiskStorage() (*Disk, error) {
	return &Disk{}, nil
}

func (d *Disk) Close() error { return nil }

func (d *Disk) Upload(ctx context.Context, r io.Reader,
	opts *gulter.UploadFileOptions,
) (*gulter.UploadedFileMetadata, error) {
	f, err := os.Create(filepath.Join(opts.Bucket,
		opts.FileName))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	return &gulter.UploadedFileMetadata{
		FolderDestination: opts.Bucket,
		Size:              n,
		Key:               opts.FileName,
	}, err
}
