package gulter_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/adelowo/gulter"
	"github.com/adelowo/gulter/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGulter(t *testing.T) {
	tt := []struct {
		name               string
		maxFileSize        int64
		pathToFile         string
		fn                 func(store *mocks.MockStorage, size int64)
		expectedStatusCode int
		validMimeTypes     []string
		// ignoreFormField instructs the test to not add the
		// multipar form data part to the request
		ignoreFormField bool
	}{
		{
			name:        "uploading succeeds",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&gulter.UploadedFileMetadata{
						Size: size,
					}, nil).
					Times(1)
			},
			expectedStatusCode: http.StatusAccepted,
			pathToFile:         "gulter.md",
			validMimeTypes:     []string{"text/markdown", "text/plain"},
		},
		{
			name:        "upload fails because form field does not exist",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&gulter.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file")).
					Times(0) // make sure this is never called
			},
			expectedStatusCode: http.StatusInternalServerError,
			pathToFile:         "gulter.md",
			validMimeTypes:     []string{"image/png", "application/pdf"},
			ignoreFormField:    true,
		},
		{
			name:        "upload fails because of mimetype validation constraints",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&gulter.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file")).
					Times(0) // make sure this is never called
			},
			expectedStatusCode: http.StatusInternalServerError,
			pathToFile:         "gulter.md",
			validMimeTypes:     []string{"image/png", "application/pdf"},
		},
		{
			name:        "upload fails because of storage layer",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&gulter.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file")).
					Times(1)
			},
			expectedStatusCode: http.StatusInternalServerError,
			pathToFile:         "gulter.md",
			validMimeTypes:     []string{"text/markdown", "text/plain"},
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := mocks.NewMockStorage(ctrl)

			handler, err := gulter.New(gulter.WithMaxFileSize(v.maxFileSize),
				gulter.WithStorage(storage),
				gulter.WithValidationFunc(gulter.MimeTypeValidator(v.validMimeTypes...)),
			)

			require.NoError(t, err)

			buffer := bytes.NewBuffer(nil)

			multipartWriter := multipart.NewWriter(buffer)

			var formFieldWriter io.Writer
			if !v.ignoreFormField {
				var err error
				formFieldWriter, err = multipartWriter.CreateFormFile("form-field", v.pathToFile)
				require.NoError(t, err)
			}

			fileToUpload, err := os.Open(filepath.Join("testdata", v.pathToFile))
			require.NoError(t, err)

			n, err := io.Copy(formFieldWriter, fileToUpload)
			require.NoError(t, err)

			v.fn(storage, int64(n))

			require.NoError(t, multipartWriter.Close())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, "/", buffer)

			r.Header.Set("Content-Type", multipartWriter.FormDataContentType())

			handler.Upload("form-field")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				file, err := gulter.FileFromContext(r, "form-field")

				require.NoError(t, err)

				require.Equal(t, v.pathToFile, file.OriginalName)

				w.WriteHeader(http.StatusAccepted)
				fmt.Fprintf(w, "successfully uploade the file")
			})).ServeHTTP(w, r)

			require.Equal(t, v.expectedStatusCode, w.Code)
		})
	}
}
