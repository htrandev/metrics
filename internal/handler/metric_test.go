package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/htrandev/metrics/internal/model"
)

var (
	errStore  = errors.New("store error")
	errGet    = errors.New("get error")
	errGetAll = errors.New("getAll error")
)

type mockStorage struct {
	storeErr bool

	getErr bool
	gauge  bool

	getAll bool
	filled bool
}

var _ MetricStorage = (*mockStorage)(nil)

func (m *mockStorage) Store(context.Context, *models.Metric) error {
	if m.storeErr {
		return errStore
	}
	return nil
}

func (m *mockStorage) Get(context.Context, string) (models.Metric, error) {
	if m.getErr {
		return models.Metric{}, errGet
	}
	metric := models.Metric{Name: "test"}
	if m.gauge {
		metric.Value.Type = models.TypeGauge
		metric.Value.Gauge = 0.1
		return metric, nil
	}

	metric.Value.Type = models.TypeCounter
	metric.Value.Counter = 1

	return metric, nil
}

func (m *mockStorage) GetAll(context.Context) ([]models.Metric, error) {
	if m.getAll {
		return nil, errGetAll
	}
	if m.filled {
		return filledStorage(), nil
	}
	return nil, nil
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

			h := NewMetricsHandler(tc.store)

			mux := http.NewServeMux()
			mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", h.Update)
			mux.ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			require.EqualValues(t, tc.expectedCode, res.StatusCode)
		})
	}
}

func TestGetHandler(t *testing.T) {
	testCases := []struct {
		name             string
		storage          *mockStorage
		method           string
		url              string
		wantErr          bool
		expectedCode     int
		expectedResponse string
	}{
		{
			name:             "valid gauge",
			storage:          &mockStorage{gauge: true},
			method:           http.MethodGet,
			url:              "/value/gauge/test",
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "0.1",
		},
		{
			name:             "valid counter",
			storage:          &mockStorage{gauge: false},
			method:           http.MethodGet,
			url:              "/value/counter/test",
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "1",
		},
		{
			name:         "unknown metric type",
			storage:      &mockStorage{},
			method:       http.MethodGet,
			url:          "/value/test/test",
			wantErr:      true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "get error",
			storage:      &mockStorage{getErr: true},
			method:       http.MethodGet,
			url:          "/value/counter/test",
			wantErr:      true,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, tc.url, nil)

			h := NewMetricsHandler(tc.storage)

			mux := http.NewServeMux()
			mux.HandleFunc("/value/{metricType}/{metricName}", h.Get)
			mux.ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			require.EqualValues(t, tc.expectedCode, res.StatusCode)
			require.EqualValues(t, tc.expectedResponse, string(body))
		})
	}
}

func TestGetAll(t *testing.T) {
	testCases := []struct {
		name             string
		storage          *mockStorage
		method           string
		wantErr          bool
		expectedCode     int
		expectedResponse string
	}{
		{
			name:             "valid empty storage",
			storage:          &mockStorage{},
			method:           http.MethodGet,
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "",
		},
		{
			name:             "valid filled storage",
			storage:          &mockStorage{filled: true},
			method:           http.MethodGet,
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "gauge: 0.1\rcounter: 1\r",
		},
		{
			name:         "get all error",
			storage:      &mockStorage{getAll: true},
			method:       http.MethodGet,
			wantErr:      true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			{
				// w := httptest.NewRecorder()
				// r := httptest.NewRequest(tc.method, tc.url, nil)

				// h := NewMetricsHandler(tc.storage)

				// mux := http.NewServeMux()
				// mux.HandleFunc("/", h.GetAll)
				// mux.ServeHTTP(w, r)

				// res := w.Result()
				// defer res.Body.Close()

				// body, err := io.ReadAll(res.Body)
				// require.NoError(t, err)

				// require.EqualValues(t, tc.expectedCode, res.StatusCode)
				// require.EqualValues(t, tc.expectedResponse, string(body))
			}

			h := NewMetricsHandler(tc.storage)
			handler := http.HandlerFunc(h.GetAll)
			srv := httptest.NewServer(handler)

			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			require.Equal(t, tc.expectedCode, resp.StatusCode())
			require.EqualValues(t, tc.expectedResponse, string(resp.Body()))
		})
	}
}

func filledStorage() []models.Metric {
	metrics := []models.Metric{
		{
			Name: "gauge",
			Value: models.MetricValue{
				Type:  models.TypeGauge,
				Gauge: 0.1,
			},
		},
		{
			Name: "counter",
			Value: models.MetricValue{
				Type:    models.TypeCounter,
				Counter: 1,
			},
		},
	}
	return metrics
}
