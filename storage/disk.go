package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adelowo/gulter"
)

type Disk struct {
	destinationFolder string
}

func NewDiskStorage(folder string) (*Disk, error) {
	if len(strings.TrimSpace(folder)) == 0 {
		return nil, errors.New("please provide a valid folder path")
	}

	return &Disk{
		destinationFolder: folder,
	}, nil
}

func (d *Disk) Close() error { return nil }

func (d *Disk) Upload(ctx context.Context, r io.Reader,
	opts *gulter.UploadFileOptions,
) (*gulter.UploadedFileMetadata, error) {
	f, err := os.Create(filepath.Join(d.destinationFolder,
		opts.FileName))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	return &gulter.UploadedFileMetadata{
		FolderDestination: d.destinationFolder,
		Size:              n,
	}, err
}
