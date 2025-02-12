package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/adelowo/gulter"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/cloudinary/cloudinary-go/v2/asset"
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

func (c *CloudinaryStore) Path(ctx context.Context,
	opts gulter.PathOptions) (string, error) {

	resp, err := c.client.Admin.Asset(ctx, admin.AssetParams{PublicID: opts.Key})
	if err != nil {
		return "", fmt.Errorf("failed to fetch asset details: %w", err)
	}

	var url *asset.Asset
	switch resp.ResourceType {
	case "image":
		url, err = c.client.Image(opts.Key)
	case "video":
		url, err = c.client.Video(opts.Key)
	case "raw":
		url, err = c.client.File(opts.Key)
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resp.ResourceType)
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate URL: %w", err)
	}

	return url.String()
}
