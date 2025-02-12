package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/adelowo/gulter"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ayinke-llc/hermes"
)

type S3Options struct {
	// If true, this will log request and responses
	DebugMode bool

	UsePathStyle bool

	// Only use if the bucket supports ACL
	ACL types.ObjectCannedACL
}

type S3Store struct {
	client *s3.Client
	opts   S3Options
}

func NewS3FromConfig(cfg aws.Config, opts S3Options) (*S3Store, error) {
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = opts.UsePathStyle

		if opts.DebugMode {
			o.ClientLogMode = aws.LogSigning | aws.LogRequest | aws.LogResponseWithBody
		}
	})

	return &S3Store{
		client: client,
		opts:   opts,
	}, nil
}

func NewS3FromEnvironment(opts S3Options) (*S3Store, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = opts.UsePathStyle

		if opts.DebugMode {
			o.ClientLogMode = aws.LogSigning | aws.LogRequest | aws.LogResponseWithBody
		}
	})

	return &S3Store{
		client: client,
		opts:   opts,
	}, nil
}

func NewS3FromClient(client *s3.Client,
	opts S3Options) (*S3Store, error) {
	return &S3Store{
		client,
		opts,
	}, nil
}

func (s *S3Store) Close() error { return nil }

func (s *S3Store) Upload(ctx context.Context, r io.Reader,
	opts *gulter.UploadFileOptions,
) (*gulter.UploadedFileMetadata, error) {

	if hermes.IsStringEmpty(opts.Bucket) {
		return nil, errors.New("please provide a valid s3 bucket")
	}

	b := new(bytes.Buffer)

	r = io.TeeReader(r, b)

	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return nil, err
	}

	seeker, err := gulter.ReaderToSeeker(b)
	if err != nil {
		return nil, err
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(opts.Bucket),
		Metadata: opts.Metadata,
		Key:      aws.String(opts.FileName),
		ACL:      s.opts.ACL,
		Body:     seeker,
	})
	if err != nil {
		return nil, err
	}

	return &gulter.UploadedFileMetadata{
		FolderDestination: opts.Bucket,
		Size:              n,
		Key:               opts.FileName,
	}, nil
}

func (s *S3Store) Path(ctx context.Context, opts gulter.PathOptions) (string, error) {

	if !opts.IsSecure {

		resp, err := s.client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: hermes.Ref(opts.Bucket),
		})
		if err != nil {
			return "", fmt.Errorf("failed to get bucket location: %w", err)
		}

		region := string(resp.LocationConstraint)
		if region == "" {
			region = "us-east-1"
		}

		url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", opts.Bucket, region, opts.Key)
		return url, nil
	}

	presignClient := s3.NewPresignClient(s.client)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: hermes.Ref(opts.Bucket),
		Key:    hermes.Ref(opts.Key),
	}, s3.WithPresignExpires(opts.ExpirationTime))
	if err != nil {
		return "", err
	}

	return presignedReq.URL, nil
}
