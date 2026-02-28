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
	"github.com/golang/mock/gomock"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/audit"
	"github.com/htrandev/metrics/internal/contracts"
	mock_contracts "github.com/htrandev/metrics/internal/contracts/mocks"
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

type mockPublisher struct{}

func (m *mockPublisher) Update(ctx context.Context, info audit.AuditInfo) {}

func TestUpdateHandler(t *testing.T) {
	ctx := context.Background()
	log := zap.NewNop()
	ctrl := gomock.NewController(t)

	testCases := []struct {
		name         string
		service      contracts.Service
		method       string
		url          string
		expectedCode int
	}{
		{
			name: "valid counter",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				m := &model.MetricDto{
					Name: "someMetric",
					Value: model.MetricValue{
						Type:    model.TypeCounter,
						Counter: 527,
					},
				}
				service.EXPECT().Store(ctx, m)
				return service
			}(),
			method:       http.MethodPost,
			url:          "/update/counter/someMetric/527",
			expectedCode: http.StatusOK,
		},
		{
			name: "valid gauge",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				m := &model.MetricDto{
					Name: "someMetric",
					Value: model.MetricValue{
						Type:  model.TypeGauge,
						Gauge: 527,
					},
				}
				service.EXPECT().Store(ctx, m)
				return service
			}(),
			method:       http.MethodPost,
			url:          "/update/gauge/someMetric/527",
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid metric type",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodPost,
			url:          "/update/test/someMetric/527",
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "set value error",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodPost,
			url:          "/update/counter/someMetric/none",
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "store error",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				m := &model.MetricDto{
					Name: "someMetric",
					Value: model.MetricValue{
						Type:    model.TypeCounter,
						Counter: 527,
					},
				}
				service.EXPECT().Store(ctx, m).Return(errStore)
				return service
			}(),
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
	ctrl := gomock.NewController(t)

	testCases := []struct {
		name             string
		service          contracts.Service
		method           string
		url              string
		wantErr          bool
		expectedCode     int
		expectedResponse string
	}{
		{
			name: "valid gauge",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().Get(gomock.Any(), "test").Return(model.MetricDto{
					Name: "test", Value: model.MetricValue{Gauge: 0.1, Type: model.TypeGauge}}, nil)
				return service
			}(),
			method:           http.MethodGet,
			url:              "/value/gauge/test",
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "0.1",
		},
		{
			name: "valid counter",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().Get(gomock.Any(), "test").Return(model.MetricDto{
					Name: "test", Value: model.MetricValue{Counter: 1, Type: model.TypeCounter}}, nil)
				return service
			}(),
			method:           http.MethodGet,
			url:              "/value/counter/test",
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "1",
		},
		{
			name: "unknown metric type",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodGet,
			url:          "/value/test/test",
			wantErr:      true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "get error",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				service.EXPECT().Get(gomock.Any(), "test").Return(model.MetricDto{}, errGet)
				return service
			}(),
			method:       http.MethodGet,
			url:          "/value/counter/test",
			wantErr:      true,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "not found",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				service.EXPECT().Get(gomock.Any(), "test").Return(model.MetricDto{}, repository.ErrNotFound)
				return service
			}(),
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
	ctrl := gomock.NewController(t)

	testCases := []struct {
		name             string
		service          contracts.Service
		method           string
		wantErr          bool
		expectedCode     int
		expectedResponse string
	}{
		{
			name: "valid empty storage",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().GetAll(gomock.Any()).Return([]model.MetricDto{}, nil)
				return service
			}(),
			method:           http.MethodGet,
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "",
		},
		{
			name: "valid filled storage",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().GetAll(gomock.Any()).Return([]model.MetricDto{
					{Name: "gauge", Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1}},
					{Name: "counter", Value: model.MetricValue{Type: model.TypeCounter, Counter: 1}},
				}, nil)
				return service
			}(),
			method:           http.MethodGet,
			wantErr:          false,
			expectedCode:     http.StatusOK,
			expectedResponse: "gauge: 0.1\rcounter: 1\r",
		},
		{
			name: "get all error",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().GetAll(gomock.Any()).Return([]model.MetricDto{}, errGetAll)
				return service
			}(),
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
	ctrl := gomock.NewController(t)

	testCases := []struct {
		name         string
		service      contracts.Service
		method       string
		expectedCode int
		body         io.Reader
	}{
		{
			name: "valid counter",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				m := &model.MetricDto{
					Name: "counter",
					Value: model.MetricValue{
						Type:    model.TypeCounter,
						Counter: 1,
					},
				}
				service.EXPECT().Store(gomock.Any(), m).Return(nil)
				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"counter","type":"counter","delta":1}`))
			}(),
		},
		{
			name: "valid gauge",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				m := &model.MetricDto{
					Name: "gauge",
					Value: model.MetricValue{
						Type:  model.TypeGauge,
						Gauge: 0.1,
					},
				}
				service.EXPECT().Store(gomock.Any(), m).Return(nil)
				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"gauge","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name: "empty counter",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"counter","type":"counter"}`))
			}(),
		},
		{
			name: "empty gauge",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"gauge","type":"gauge"}`))
			}(),
		},
		{
			name: "build request error",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id:"gauge","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name: "empty name",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusNotFound,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`{"id":"","type":"gauge","value":0.1}`))
			}(),
		},
		{
			name: "store error",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				m := &model.MetricDto{
					Name: "gauge",
					Value: model.MetricValue{
						Type:  model.TypeGauge,
						Gauge: 0.1,
					},
				}
				service.EXPECT().Store(gomock.Any(), m).Return(errStore)
				return service
			}(),
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
	ctrl := gomock.NewController(t)

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
		service      contracts.Service
		method       string
		expectedCode int
		body         io.Reader
	}{
		{
			name: "valid",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				m := []model.MetricDto{
					{Name: "gauge", Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1}},
					{Name: "counter", Value: model.MetricValue{Type: model.TypeCounter, Counter: 1}},
				}
				service.EXPECT().StoreManyWithRetry(gomock.Any(), m).Return(nil)
				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				b, err := easyjson.Marshal(metrics)
				require.NoError(t, err)

				return bytes.NewBuffer(b)
			}(),
		},
		{
			name: "invalid body",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(``))
			}(),
		},
		{
			name: "empty metrics",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
			method:       http.MethodPost,
			expectedCode: http.StatusOK,
			body: func() io.Reader {
				return bytes.NewBuffer([]byte(`[]`))
			}(),
		},
		{
			name: "store error",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				m := []model.MetricDto{
					{Name: "gauge", Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1}},
					{Name: "counter", Value: model.MetricValue{Type: model.TypeCounter, Counter: 1}},
				}
				service.EXPECT().StoreManyWithRetry(gomock.Any(), m).Return(errStoreMany)
				return service
			}(),
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
	ctrl := gomock.NewController(t)

	testCases := []struct {
		name         string
		service      contracts.Service
		method       string
		expectedCode int
		body         io.Reader
		expectedBody string
	}{
		{
			name: "valid gauge",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				service.EXPECT().Get(gomock.Any(), "test").Return(model.MetricDto{Name: "test", Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1}}, nil)
				return service
			}(),
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
			name: "valid counter",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				service.EXPECT().Get(gomock.Any(), "test").Return(model.MetricDto{Name: "test", Value: model.MetricValue{Type: model.TypeCounter, Counter: 1}}, nil)
				return service
			}(),
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
			name: "unknown type",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
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
			name: "error build request",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
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
			name: "empty name",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)

				return service
			}(),
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
			name: "get error",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().Get(gomock.Any(), "test").Return(model.MetricDto{}, errGet)
				return service
			}(),
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
			name: "not found",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().Get(gomock.Any(), "test").Return(model.MetricDto{}, repository.ErrNotFound)
				return service
			}(),
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
	ctrl := gomock.NewController(t)

	testCases := []struct {
		name         string
		service      contracts.Service
		expectedCode int
	}{
		{
			name: "valid",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().Ping(gomock.Any()).Return(nil)
				return service
			}(),
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid",
			service: func() contracts.Service {
				service := mock_contracts.NewMockService(ctrl)
				service.EXPECT().Ping(gomock.Any()).Return(errPing)
				return service
			}(),
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
