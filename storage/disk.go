package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/adelowo/gulter"
	"github.com/ayinke-llc/hermes"
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

	if hermes.IsStringEmpty(opts.Bucket) {
		return nil, errors.New("please provide a valid folder")
	}

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
