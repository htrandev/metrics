package handler

import (
	"bytes"
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
	"github.com/htrandev/metrics/pkg/logger"
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
	log, err := logger.NewZapLogger("debug")
	require.NoError(t, err)

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

			h := NewMetricsHandler(log, tc.store)

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
	log, err := logger.NewZapLogger("debug")
	require.NoError(t, err)

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

			h := NewMetricsHandler(log, tc.storage)

			mux := http.NewServeMux()
			mux.HandleFunc("/value/{metricType}/{metricName}", h.Get)
			mux.ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			require.EqualValues(t, tc.expectedCode, res.StatusCode)
			require.EqualValues(t, tc.expectedResponse, string(body))

			// h := NewMetricsHandler(log, tc.storage)
			// handler := http.HandlerFunc(h.Get)
			// srv := httptest.NewServer(handler)

			// req := resty.New().R()
			// req.Method = tc.method
			// log.Info(srv.URL + tc.url)
			// req.URL = srv.URL + tc.url

			// resp, err := req.Send()
			// assert.NoError(t, err, "error making HTTP request")

			// require.EqualValues(t, tc.expectedCode, resp.StatusCode())
			// require.EqualValues(t, tc.expectedResponse, string(resp.Body()))
		})
	}
}

func TestGetAll(t *testing.T) {
	log, err := logger.NewZapLogger("debug")
	require.NoError(t, err)

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
			h := NewMetricsHandler(log, tc.storage)
			handler := http.HandlerFunc(h.GetAll)
			srv := httptest.NewServer(handler)
			defer srv.Close()

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

func TestUpdateViaBody(t *testing.T) {
	log, err := logger.NewZapLogger("debug")
	require.NoError(t, err)

	testCases := []struct {
		name         string
		storage      *mockStorage
		method       string
		expectedCode int
		body         io.Reader
	}{
		{
			name:         "valid counter",
			storage:      &mockStorage{},
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"counter","type":"counter","value":1}`))
			}(),
		},
		{
			name:         "valid gauge",
			storage:      &mockStorage{},
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"gauge","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name:         "build request error",
			storage:      &mockStorage{},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id:"gauge","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name:         "empty name",
			storage:      &mockStorage{},
			method:       http.MethodPost,
			expectedCode: http.StatusNotFound,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name:         "store error",
			storage:      &mockStorage{storeErr: true},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"gauge","type":"gauge","value":0.1}`))
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewMetricsHandler(log, tc.storage)
			handler := http.HandlerFunc(h.UpdateViaBody)
			srv := httptest.NewServer(handler)
			defer srv.Close()

			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL
			req.Body = tc.body

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			require.EqualValues(t, tc.expectedCode, resp.StatusCode())
		})
	}
}

func TestGetViaBody(t *testing.T) {
	log, err := logger.NewZapLogger("debug")
	require.NoError(t, err)

	testCases := []struct {
		name         string
		storage      *mockStorage
		method       string
		expectedCode int
		body         io.Reader
		expectedBody string
	}{
		{
			name:         "valid gauge",
			storage:      &mockStorage{gauge: true},
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"test","type":"gauge"}`))
			}(),
			expectedBody: func() string {
				return `{"id":"test","type":"gauge","value":0.1}`
			}(),
		},
		{
			name:         "valid counter",
			storage:      &mockStorage{gauge: false},
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"test","type":"counter"}`))
			}(),
			expectedBody: func() string {
				return `{"id":"test","type":"counter","delta":1}`
			}(),
		},
		{
			name:         "unknown type",
			storage:      &mockStorage{gauge: false},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"test","type":"test"}`))
			}(),
			expectedBody: func() string {
				return ``
			}(),
		},
		{
			name:         "error build request",
			storage:      &mockStorage{gauge: false},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id:"counter","type":"counter"}`))
			}(),
			expectedBody: func() string {
				return ``
			}(),
		},
		{
			name:         "empty name",
			storage:      &mockStorage{gauge: false},
			method:       http.MethodPost,
			expectedCode: http.StatusNotFound,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"","type":"counter"}`))
			}(),
			expectedBody: func() string {
				return ``
			}(),
		},
		{
			name:         "get error",
			storage:      &mockStorage{getErr: true},
			method:       http.MethodPost,
			expectedCode: http.StatusNotFound,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"test","type":"counter"}`))
			}(),
			expectedBody: func() string {
				return ``
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewMetricsHandler(log, tc.storage)
			handler := http.HandlerFunc(h.GetViaBody)
			srv := httptest.NewServer(handler)
			defer srv.Close()

			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL
			req.Body = tc.body

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			require.EqualValues(t, tc.expectedCode, resp.StatusCode())
			require.EqualValues(t, tc.expectedBody, string(resp.Body()))
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
