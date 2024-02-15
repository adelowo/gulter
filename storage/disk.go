package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type Disk struct {
	destinationFolder string
}

func (d *Disk) Close() error { return nil }

func (d *Disk) Upload(ctx context.Context, r io.Reader,
	fileName string,
) error {
	f, err := os.Create(filepath.Join(d.destinationFolder, fileName))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}
