package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMethodChecker(t *testing.T) {
	dummyHandler := dummyHandler()

	testCases := []struct {
		name         string
		method       string
		req          *http.Request
		expectedCode int
	}{
		{
			name:         "valid",
			method:       http.MethodGet,
			req:          httptest.NewRequest(http.MethodGet, "/", nil),
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid method",
			method:       http.MethodGet,
			req:          httptest.NewRequest(http.MethodPost, "/", nil),
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			wrapper := MethodChecker(tc.method)
			wrapper(dummyHandler).ServeHTTP(rec, tc.req)

			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}
