package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContentType(t *testing.T) {
	dummyHandler := dummyHandler()

	testCases := []struct {
		name        string
		contentType string
		statusCode  int
	}{
		{
			name:        "valid",
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "invalid",
			contentType: "text/plain",
			statusCode:  http.StatusNotAcceptable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			req.Header.Set("Content-Type", tc.contentType)

			wrapper := ContentType()
			wrapper(dummyHandler).ServeHTTP(rec, req)

			require.Equal(t, tc.statusCode, rec.Code)

		})
	}
}
