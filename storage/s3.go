package storage

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/adelowo/gulter"
	"github.com/adelowo/gulter/internal/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Options struct {
	Bucket string
	ACL    types.ObjectCannedACL
}

type S3Store struct {
	client *s3.Client
	opts   S3Options
}

func NewS3FromConfig(cfg aws.Config, opts S3Options) (*S3Store, error) {
	if util.IsStringEmpty(opts.Bucket) {
		return nil, errors.New("please provide a valid s3 bucket")
	}

	return &S3Store{
		client: s3.NewFromConfig(cfg),
		opts:   opts,
	}, nil
}

func NewS3FromEnvironment(opts S3Options) (*S3Store, error) {
	if util.IsStringEmpty(opts.Bucket) {
		return nil, errors.New("please provide a valid s3 bucket")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return &S3Store{
		client: s3.NewFromConfig(cfg),
		opts:   opts,
	}, nil
}

func NewS3FromClient(client *s3.Client, opts S3Options) (*S3Store, error) {
	if util.IsStringEmpty(opts.Bucket) {
		return nil, errors.New("please provide a valid s3 bucket")
	}
	return &S3Store{
		client,
		opts,
	}, nil
}

func (s *S3Store) Close() error { return nil }

func (s *S3Store) Upload(ctx context.Context, r io.Reader,
	opts *gulter.UploadFileOptions,
) (*gulter.UploadedFileMetadata, error) {
	b := new(bytes.Buffer)

	r = io.TeeReader(r, b)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(s.opts.Bucket),
		Metadata: opts.Metadata,
		Key:      aws.String(opts.FileName),
		ACL:      s.opts.ACL,
	})
	if err != nil {
		return nil, err
	}

	return &gulter.UploadedFileMetadata{
		FolderDestination: s.opts.Bucket,
		Size:              int64(b.Len()),
	}, nil
}
