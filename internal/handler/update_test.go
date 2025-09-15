package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/htrandev/metrics/internal/model"
	"github.com/stretchr/testify/require"
)

var errStore = errors.New("store error")

type mockStorage struct {
	storeErr bool
}

var _ MetricStorage = (*mockStorage)(nil)

func (m *mockStorage) Store(_ *models.Metric) error {
	if m.storeErr {
		return errStore
	}
	return nil
}

func TestUpdateHandler(t *testing.T) {
	testCases := []struct {
		name         string
		store        *mockStorage
		method       string
		url          string
		expectedCode int
	}{
		{
			name:         "valid counter",
			store:        &mockStorage{},
			method:       http.MethodPost,
			url:          "/update/counter/someMetric/527",
			expectedCode: http.StatusOK,
		},
		{
			name:         "valid gauge",
			store:        &mockStorage{},
			method:       http.MethodPost,
			url:          "/update/gauge/someMetric/527",
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid method",
			store:        &mockStorage{},
			method:       http.MethodGet,
			url:          "/update/counter/someMetric/527",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "invalid metric type",
			store:        &mockStorage{},
			method:       http.MethodPost,
			url:          "/update/test/someMetric/527",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "set value error",
			store:        &mockStorage{storeErr: true},
			method:       http.MethodPost,
			url:          "/update/counter/someMetric/none",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "store error",
			store:        &mockStorage{storeErr: true},
			method:       http.MethodPost,
			url:          "/update/counter/someMetric/527",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, tc.url, nil)

			h := NewUpdateHandler(tc.store)

			mux := http.NewServeMux()
			mux.Handle("/update/{metricType}/{metricName}/{metricValue}", h)
			mux.ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			require.EqualValues(t, tc.expectedCode, res.StatusCode)
		})
	}
}
