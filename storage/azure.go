package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
	"github.com/adelowo/gulter"
	"github.com/ayinke-llc/hermes"
)

const (
	// sasTimeFormat is the ISO 8601 UTC layout Azure expects for user delegation key times
	sasTimeFormat = "2006-01-02T15:04:05Z"

	// defaultAzureSASExpiry is used when PathOptions does not provide an expiration
	defaultAzureSASExpiry = 15 * time.Minute
)

type AzureBlobOptions struct {
	ContainerName string
	ServiceUrl    string

	// If true, this will log request and responses
	DebugMode bool

	// Allows all types of azcore TokenCredential interface children
	// i.e azcore.DefaultAzureCredential, azcore.NewManagedIdentityCredential
	// azcore.NewWorkloadIdentityCredential or azcore.NewClientSecretCredential
	Credential azcore.TokenCredential

	// SharedKey should be set when the client is built from a storage account
	// key ( NewAzureBlobStoreFromClient with a shared key client ). It lets
	// Path sign account key SAS urls directly. Without it, a secure Path falls
	// back to a user delegation key which only works with an Azure AD credential
	SharedKey *azblob.SharedKeyCredential

	UploadStreamOpts azblob.UploadStreamOptions

	StorageAccount string
}

type AzureBlobStore struct {
	client        *azblob.Client
	opts          AzureBlobOptions
	containerName string
}

// azureServiceURL resolves the blob service url. It prefers an explicitly
// provided ServiceUrl (needed for custom domains, sovereign clouds or the
// local azurite emulator) and falls back to deriving one from the storage
// account name for the public cloud
func azureServiceURL(opts AzureBlobOptions) (string, error) {
	if !hermes.IsStringEmpty(opts.ServiceUrl) {
		return opts.ServiceUrl, nil
	}

	if hermes.IsStringEmpty(opts.StorageAccount) {
		return "", errors.New("please provide a valid azure blob service url or storage account")
	}

	return fmt.Sprintf("https://%s.blob.core.windows.net/", opts.StorageAccount), nil
}

func azureClientOptions(debugMode bool) *azblob.ClientOptions {
	logOpts := policy.LogOptions{}
	if debugMode {
		logOpts.IncludeBody = true
	}

	return &azblob.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Logging: logOpts,
		},
	}
}

func NewAzureBlobStoreFromClient(client *azblob.Client, opts AzureBlobOptions) (*AzureBlobStore, error) {
	if hermes.IsStringEmpty(opts.ContainerName) {
		return nil, errors.New("please provide a valid azure container name")
	}

	if client == nil {
		return nil, errors.New("please provide a valid azure blob client")
	}

	return &AzureBlobStore{
		client:        client,
		opts:          opts,
		containerName: opts.ContainerName,
	}, nil
}

func NewAzureBlobFromCredential(opts AzureBlobOptions) (*AzureBlobStore, error) {
	if hermes.IsStringEmpty(opts.ContainerName) {
		return nil, errors.New("please provide a valid azure container name")
	}

	serviceURL, err := azureServiceURL(opts)
	if err != nil {
		return nil, err
	}
	opts.ServiceUrl = serviceURL

	if opts.Credential == nil {
		credential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}

		opts.Credential = credential
	}

	client, err := azblob.NewClient(opts.ServiceUrl, opts.Credential, azureClientOptions(opts.DebugMode))
	if err != nil {
		return nil, err
	}

	return &AzureBlobStore{
		client:        client,
		opts:          opts,
		containerName: opts.ContainerName,
	}, nil
}

func (a *AzureBlobStore) Close() error {
	return nil
}

func (a *AzureBlobStore) Upload(ctx context.Context, r io.Reader, opts *gulter.UploadFileOptions) (*gulter.UploadedFileMetadata, error) {
	uploadOpts := a.azureUploadStreamOptions(opts)

	reader := &countingReader{Reader: r}

	_, err := a.client.UploadStream(ctx, a.containerName, opts.FileName, reader, uploadOpts)
	if err != nil {
		return nil, err
	}

	return &gulter.UploadedFileMetadata{
		FolderDestination: a.containerName,
		Size:              reader.bytesRead,
		Key:               opts.FileName,
	}, nil
}

func (a *AzureBlobStore) azureUploadStreamOptions(opts *gulter.UploadFileOptions) *azblob.UploadStreamOptions {
	uploadOpts := a.opts.UploadStreamOpts

	if len(opts.Metadata) > 0 {
		uploadOpts.Metadata = azureBlobMetadata(uploadOpts.Metadata, opts.Metadata)
	}

	if opts.ContentType != "" {
		headers := blob.HTTPHeaders{}
		if uploadOpts.HTTPHeaders != nil {
			headers = *uploadOpts.HTTPHeaders
		}

		headers.BlobContentType = hermes.Ref(opts.ContentType)
		uploadOpts.HTTPHeaders = &headers
	}

	return &uploadOpts
}

type countingReader struct {
	io.Reader
	bytesRead int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.bytesRead += int64(n)
	return n, err
}

func azureBlobMetadata(base map[string]*string, metadata map[string]string) map[string]*string {
	merged := make(map[string]*string, len(base)+len(metadata))
	for key, value := range base {
		merged[key] = value
	}

	for key, value := range metadata {
		merged[key] = hermes.Ref(value)
	}

	return merged
}

func (a *AzureBlobStore) Path(ctx context.Context, opts gulter.PathOptions) (string, error) {
	base := strings.TrimRight(a.client.ServiceClient().URL(), "/")
	blobURL := fmt.Sprintf("%s/%s/%s", base, a.containerName, opts.Key)

	if !opts.IsSecure {
		return blobURL, nil
	}

	expiry := opts.ExpirationTime
	if expiry <= 0 {
		expiry = defaultAzureSASExpiry
	}

	now := time.Now().UTC()
	end := now.Add(expiry)

	signatureValues := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     now,
		ExpiryTime:    end,
		Permissions:   (&sas.BlobPermissions{Read: true}).String(),
		ContainerName: a.containerName,
		BlobName:      opts.Key,
	}

	var (
		queryParams sas.QueryParameters
		err         error
	)

	// an account key can sign the SAS directly. otherwise fall back to a user
	// delegation key, which requires the client to be backed by an Azure AD
	// credential
	if a.opts.SharedKey != nil {
		queryParams, err = signatureValues.SignWithSharedKey(a.opts.SharedKey)
	} else {
		var udc *service.UserDelegationCredential
		udc, err = a.client.ServiceClient().GetUserDelegationCredential(ctx, service.KeyInfo{
			Start:  hermes.Ref(now.Format(sasTimeFormat)),
			Expiry: hermes.Ref(end.Format(sasTimeFormat)),
		}, nil)
		if err != nil {
			return "", fmt.Errorf("failed to get user delegation credential: %w", err)
		}

		queryParams, err = signatureValues.SignWithUserDelegation(udc)
	}
	if err != nil {
		return "", fmt.Errorf("failed to sign blob sas: %w", err)
	}

	return blobURL + "?" + queryParams.Encode(), nil
}
