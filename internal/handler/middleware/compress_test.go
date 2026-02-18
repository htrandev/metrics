package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCompress(t *testing.T) {
	const (
		acceptEncoding  string = "Accept-Encoding"
		contentEncoding string = "Content-Encoding"
	)

	logger := zap.NewNop()

	body := `{"test": "compress"}`

	dummyHandler := dummyHandler()
	testCases := []struct {
		name                    string
		req                     *http.Request
		expectedStatus          int
		expectedContentEncoding string
	}{
		{
			name: "support gzip",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/", nil)
				r.Header.Set(acceptEncoding, "gzip")
				return r
			}(),
			expectedStatus:          http.StatusOK,
			expectedContentEncoding: "gzip",
		},
		{
			name: "sends gzip",
			req: func() *http.Request {
				buf := bytes.NewBuffer(nil)
				zb := gzip.NewWriter(buf)
				_, err := zb.Write([]byte(body))
				require.NoError(t, err)
				err = zb.Close()
				require.NoError(t, err)

				r := httptest.NewRequest(http.MethodPost, "/", buf)
				r.Header.Set(contentEncoding, "gzip")
				return r
			}(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "sends gzip error",
			req: func() *http.Request {
				buf := bytes.NewBuffer(nil)

				r := httptest.NewRequest(http.MethodPost, "/", buf)
				r.Header.Set(contentEncoding, "gzip")
				return r
			}(),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			wrapper := Compress(logger)
			wrapper(dummyHandler).ServeHTTP(rec, tc.req)

			require.Equal(t, tc.expectedStatus, rec.Code)
			require.Equal(t, tc.expectedContentEncoding, rec.Header().Get(contentEncoding))
		})
	}
}
