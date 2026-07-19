package storage

import (
	"io"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/adelowo/gulter"
	"github.com/stretchr/testify/require"
)

func TestAzureUploadStreamOptions(t *testing.T) {
	defaultContentType := "application/octet-stream"
	defaultMetadataValue := "default"

	store := &AzureBlobStore{
		opts: AzureBlobOptions{
			UploadStreamOpts: azblob.UploadStreamOptions{
				HTTPHeaders: &blob.HTTPHeaders{
					BlobContentType: &defaultContentType,
				},
				Metadata: map[string]*string{
					"default": &defaultMetadataValue,
				},
			},
		},
	}

	uploadOpts := store.azureUploadStreamOptions(&gulter.UploadFileOptions{
		ContentType: "image/jpeg",
		Metadata: map[string]string{
			"owner": "gulter",
		},
	})

	require.Equal(t, "image/jpeg", *uploadOpts.HTTPHeaders.BlobContentType)
	require.Equal(t, "default", *uploadOpts.Metadata["default"])
	require.Equal(t, "gulter", *uploadOpts.Metadata["owner"])

	require.Equal(t, "application/octet-stream", *store.opts.UploadStreamOpts.HTTPHeaders.BlobContentType)
	require.NotContains(t, store.opts.UploadStreamOpts.Metadata, "owner")
}

func TestCountingReader(t *testing.T) {
	reader := &countingReader{Reader: strings.NewReader("hello")}

	_, err := io.Copy(io.Discard, reader)
	require.NoError(t, err)
	require.Equal(t, int64(5), reader.bytesRead)
}
