package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adelowo/gulter"
	"github.com/ayinke-llc/hermes"
)

type Disk struct {
	folder string
}

func NewDiskStorage(pathToFolder string) (*Disk, error) {
	if hermes.IsStringEmpty(pathToFolder) {
		return nil, errors.New("please provide a bucket")
	}

	if _, err := os.Stat(pathToFolder); err != nil {
		return nil, err
	}

	return &Disk{
		folder: pathToFolder,
	}, nil
}

func (d *Disk) Close() error { return nil }

func (d *Disk) Upload(ctx context.Context, r io.Reader,
	opts *gulter.UploadFileOptions,
) (*gulter.UploadedFileMetadata, error) {

	f, err := os.Create(filepath.Join(d.folder, opts.FileName))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return nil, err
	}

	return &gulter.UploadedFileMetadata{
		FolderDestination: d.folder,
		Size:              n,
		Key:               opts.FileName,
	}, err
}

func (d *Disk) Path(ctx context.Context,
	opts gulter.PathOptions) (string, error) {
	return fmt.Sprintf("%s/%s", d.folder, opts.Key), nil
}
