package gulter_test

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
		fn                 func(store *mocks.MockStorage, size int64)
		expectedStatusCode int
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
		},
		{
			name:        "upload fails",
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
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := mocks.NewMockStorage(ctrl)

			handler, err := gulter.New(gulter.WithMaxFileSize(v.maxFileSize),
				gulter.WithStorage(storage))

			require.NoError(t, err)

			buffer := bytes.NewBuffer(nil)

			multipartWriter := multipart.NewWriter(buffer)

			f, err := multipartWriter.CreateFormFile("form-field", "gulter.txt")
			require.NoError(t, err)

			n, err := f.Write([]byte(`lanre is working on something`))
			require.NoError(t, err)

			v.fn(storage, int64(n))

			require.NoError(t, multipartWriter.Close())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, "/", buffer)

			r.Header.Set("Content-Type", multipartWriter.FormDataContentType())

			handler.Upload("form-field")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				file, err := gulter.FileFromContext(r, "form-field")

				require.NoError(t, err)

				require.Equal(t, "gulter.txt", file.OriginalName)

				w.WriteHeader(http.StatusAccepted)
				fmt.Fprintf(w, "successfully uploade the file")
			})).ServeHTTP(w, r)

			require.Equal(t, v.expectedStatusCode, w.Code)
		})
	}
}
