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
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/audit"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/repository"
)

var (
	errStore     = errors.New("store error")
	errStoreMany = errors.New("store many error")
	errGet       = errors.New("get error")
	errGetAll    = errors.New("getAll error")
	errPing      = errors.New("ping error")
)

type mockService struct {
	storeErr     bool
	storeManyErr bool
	pingErr      bool

	notFound bool
	getErr   bool
	gauge    bool

	getAll bool
	filled bool
}

var _ Service = (*mockService)(nil)

func (m *mockService) Store(context.Context, *model.Metric) error {
	if m.storeErr {
		return errStore
	}
	return nil
}

func (m *mockService) StoreMany(context.Context, []model.Metric) error {
	if m.storeManyErr {
		return errStoreMany
	}
	return nil
}

func (m *mockService) StoreManyWithRetry(context.Context, []model.Metric) error {
	if m.storeManyErr {
		return errStoreMany
	}
	return nil
}

func (m *mockService) Get(context.Context, string) (model.Metric, error) {
	if m.notFound {
		return model.Metric{}, repository.ErrNotFound
	}
	if m.getErr {
		return model.Metric{}, errGet
	}
	metric := model.Metric{Name: "test"}
	if m.gauge {
		metric.Value.Type = model.TypeGauge
		metric.Value.Gauge = 0.1
		return metric, nil
	}

	metric.Value.Type = model.TypeCounter
	metric.Value.Counter = 1

	return metric, nil
}

func (m *mockService) GetAll(context.Context) ([]model.Metric, error) {
	if m.getAll {
		return nil, errGetAll
	}
	if m.filled {
		return filledStorage(), nil
	}
	return nil, nil
}

func (m *mockService) Ping(context.Context) error {
	if m.pingErr {
		return errPing
	}
	return nil
}

type mockPublisher struct{}

func (m *mockPublisher) Update(ctx context.Context, info audit.AuditInfo) {}

func TestUpdateHandler(t *testing.T) {
	log := zap.NewNop()

	testCases := []struct {
		name         string
		service      *mockService
		method       string
		url          string
		expectedCode int
	}{
		{
			name:         "valid counter",
			service:      &mockService{},
			method:       http.MethodPost,
			url:          "/update/counter/someMetric/527",
			expectedCode: http.StatusOK,
		},
		{
			name:         "valid gauge",
			service:      &mockService{},
			method:       http.MethodPost,
			url:          "/update/gauge/someMetric/527",
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid metric type",
			service:      &mockService{},
			method:       http.MethodPost,
			url:          "/update/test/someMetric/527",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "set value error",
			service:      &mockService{storeErr: true},
			method:       http.MethodPost,
			url:          "/update/counter/someMetric/none",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "store error",
			service:      &mockService{storeErr: true},
			method:       http.MethodPost,
			url:          "/update/counter/someMetric/527",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, tc.url, nil)

			h := NewMetricsHandler(
				log,
				tc.service,
				&mockPublisher{},
			)

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
	log := zap.NewNop()

	testCases := []struct {
		name             string
		service          *mockService
		method           string
		url              string
		wantErr          bool
		expectedCode     int
		expectedResponse string
	}{
		{
			name:             "valid gauge",
			service:          &mockService{gauge: true},
			method:           http.MethodGet,
			url:              "/value/gauge/test",
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "0.1",
		},
		{
			name:             "valid counter",
			service:          &mockService{gauge: false},
			method:           http.MethodGet,
			url:              "/value/counter/test",
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "1",
		},
		{
			name:         "unknown metric type",
			service:      &mockService{},
			method:       http.MethodGet,
			url:          "/value/test/test",
			wantErr:      true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "get error",
			service:      &mockService{getErr: true},
			method:       http.MethodGet,
			url:          "/value/counter/test",
			wantErr:      true,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "not found",
			service:      &mockService{notFound: true},
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

			h := NewMetricsHandler(
				log,
				tc.service,
				&mockPublisher{},
			)

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
	log := zap.NewNop()

	testCases := []struct {
		name             string
		service          *mockService
		method           string
		wantErr          bool
		expectedCode     int
		expectedResponse string
	}{
		{
			name:             "valid empty storage",
			service:          &mockService{},
			method:           http.MethodGet,
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "",
		},
		{
			name:             "valid filled storage",
			service:          &mockService{filled: true},
			method:           http.MethodGet,
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "gauge: 0.1\rcounter: 1\r",
		},
		{
			name:         "get all error",
			service:      &mockService{getAll: true},
			method:       http.MethodGet,
			wantErr:      true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewMetricsHandler(
				log,
				tc.service,
				&mockPublisher{},
			)
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

func TestUpdateJSON(t *testing.T) {
	log := zap.NewNop()

	testCases := []struct {
		name         string
		service      *mockService
		method       string
		expectedCode int
		body         io.Reader
	}{
		{
			name:         "valid counter",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"counter","type":"counter","delta":1}`))
			}(),
		},
		{
			name:         "valid gauge",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"gauge","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name:         "empty counter",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"counter","type":"counter"}`))
			}(),
		},
		{
			name:         "empty gauge",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"gauge","type":"gauge"}`))
			}(),
		},
		{
			name:         "build request error",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id:"gauge","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name:         "empty name",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusNotFound,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name:         "store error",
			service:      &mockService{storeErr: true},
			method:       http.MethodPost,
			expectedCode: http.StatusInternalServerError,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"gauge","type":"gauge","value":0.1}`))
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewMetricsHandler(
				log,
				tc.service,
				&mockPublisher{},
			)
			handler := http.HandlerFunc(h.UpdateJSON)
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

func TestUpdateManyJSON(t *testing.T) {
	log := zap.NewNop()

	var (
		delta   int64 = 1
		value         = 0.1
		metrics       = model.MetricsSlice{
			{
				ID:    "gauge",
				MType: "gauge",
				Value: &value,
			},
			{
				ID:    "counter",
				MType: "counter",
				Delta: &delta,
			},
		}
	)

	testCases := []struct {
		name         string
		service      *mockService
		method       string
		expectedCode int
		body         io.Reader
	}{
		{
			name:         "valid",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				b, err := easyjson.Marshal(metrics)
				require.NoError(t, err)

				return bytes.NewBuffer(b)
			}(),
		},
		{
			name:         "invalid body",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(``))
			}(),
		},
		{
			name:         "empty metrics",
			service:      &mockService{},
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`[]`))
			}(),
		},
		{
			name:         "store error",
			service:      &mockService{storeManyErr: true},
			method:       http.MethodPost,
			expectedCode: http.StatusInternalServerError,
			body: func() io.Reader {
				b, err := easyjson.Marshal(metrics)
				require.NoError(t, err)

				return bytes.NewBuffer(b)
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewMetricsHandler(
				log,
				tc.service,
				&mockPublisher{},
			)
			handler := http.HandlerFunc(h.UpdateManyJSON)
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

func TestGetJSON(t *testing.T) {
	log := zap.NewNop()

	testCases := []struct {
		name         string
		service      *mockService
		method       string
		expectedCode int
		body         io.Reader
		expectedBody string
	}{
		{
			name:         "valid gauge",
			service:      &mockService{gauge: true},
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
			service:      &mockService{gauge: false},
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
			service:      &mockService{gauge: false},
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
			service:      &mockService{gauge: false},
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
			service:      &mockService{gauge: false},
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"","type":"counter"}`))
			}(),
			expectedBody: func() string {
				return ``
			}(),
		},
		{
			name:         "get error",
			service:      &mockService{getErr: true},
			method:       http.MethodPost,
			expectedCode: http.StatusInternalServerError,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"test","type":"counter"}`))
			}(),
			expectedBody: func() string {
				return ``
			}(),
		},
		{
			name:         "not found",
			service:      &mockService{notFound: true},
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
			h := NewMetricsHandler(
				log,
				tc.service,
				&mockPublisher{},
			)
			handler := http.HandlerFunc(h.GetJSON)
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

func TestPing(t *testing.T) {
	log := zap.NewNop()

	testCases := []struct {
		name         string
		service      *mockService
		expectedCode int
	}{
		{
			name:         "valid",
			service:      &mockService{},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid",
			service:      &mockService{pingErr: true},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewMetricsHandler(
				log,
				tc.service,
				&mockPublisher{},
			)
			handler := http.HandlerFunc(h.Ping)
			srv := httptest.NewServer(handler)
			defer srv.Close()

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = srv.URL

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			require.EqualValues(t, tc.expectedCode, resp.StatusCode())
		})
	}

}

func filledStorage() []model.Metric {
	metrics := []model.Metric{
		{
			Name: "gauge",
			Value: model.MetricValue{
				Type:  model.TypeGauge,
				Gauge: 0.1,
			},
		},
		{
			Name: "counter",
			Value: model.MetricValue{
				Type:    model.TypeCounter,
				Counter: 1,
			},
		},
	}
	return metrics
}
