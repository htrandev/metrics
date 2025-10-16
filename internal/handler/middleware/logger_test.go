package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/htrandev/metrics/pkg/logger"
)

func TestLogger(t *testing.T) {
	testCases := []struct {
		name       string
		handler    http.HandlerFunc
		request    *http.Request
		statusCode int
		method     string
		path       string
		size       int
	}{
		{
			name:       "valid get",
			handler:    dummyHandler(),
			request:    httptest.NewRequest(http.MethodGet, "/value/counter/testMetrics", nil),
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			path:       "/value/counter/testMetrics",
			size:       0,
		},
		{
			name:       "valid post",
			handler:    dummyHandlerWithResponse(),
			request:    httptest.NewRequest(http.MethodPost, "/update/counter/testMetric/100", nil),
			statusCode: http.StatusOK,
			method:     http.MethodPost,
			path:       "/update/counter/testMetric/100",
			size:       6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			lm, err := logger.NewZapLogger("info")
			require.NoError(t, err)

			wrapper := Logger(lm)
			wrapper(tc.handler).ServeHTTP(rec, tc.request)

			require.Equal(t, tc.statusCode, rec.Code)
			require.Equal(t, tc.method, tc.request.Method)
			require.Equal(t, tc.path, tc.request.URL.RequestURI())
			require.Equal(t, tc.size, rec.Body.Len())
		})
	}
}
