package gulter

import (
	"io"
	"os"
)

func ReaderToSeeker(r io.Reader) (io.ReadSeeker, error) {
	tmpfile, err := os.CreateTemp("", "upload-")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmpfile, r)
	if err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())
		return nil, err
	}

	_, err = tmpfile.Seek(0, 0)
	if err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())
		return nil, err
	}

	// Return the file, which implements io.ReadSeeker
	// which you can now pass to the gulter uploader
	return tmpfile, nil
}
