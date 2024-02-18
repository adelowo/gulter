package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/adelowo/gulter"
)

type Disk struct {
	destinationFolder string
}

func (d *Disk) Close() error { return nil }

func (d *Disk) Upload(ctx context.Context, r io.Reader,
	opts *gulter.UploadFileOptions,
) (*gulter.File, error) {
	f, err := os.Create(filepath.Join(d.destinationFolder,
		opts.FileName))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return nil, err
}
