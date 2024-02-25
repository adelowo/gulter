package storage

import (
	"context"
	"io"

	"github.com/adelowo/gulter"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryOptions struct {
	UniqueFilesOnly       bool
	OverwriteExistingFile bool
	CloudName             string
	APIKey                string
	APISecret             string
}

type CloudinaryStore struct {
	client *cloudinary.Cloudinary
	opts   CloudinaryOptions
}

func NewCloudinary(opts CloudinaryOptions) (*CloudinaryStore, error) {
	client, err := cloudinary.NewFromParams(
		opts.CloudName, opts.APIKey,
		opts.APISecret,
	)
	if err != nil {
		return nil, err
	}

	return &CloudinaryStore{
		client: client,
		opts:   opts,
	}, nil
}

func (c *CloudinaryStore) Close() error { return nil }

func (c *CloudinaryStore) Upload(ctx context.Context,
	r io.Reader, opts *gulter.UploadFileOptions,
) (*gulter.UploadedFileMetadata, error) {
	resp, err := c.client.Upload.Upload(ctx,
		r, uploader.UploadParams{
			PublicID:       opts.FileName,
			UniqueFilename: api.Bool(c.opts.UniqueFilesOnly),
			Overwrite:      api.Bool(c.opts.OverwriteExistingFile),
		})
	if err != nil {
		return nil, err
	}

	return &gulter.UploadedFileMetadata{
		FolderDestination: "",
		Size:              int64(resp.Bytes),
		Key:               resp.PublicID,
	}, nil
}
