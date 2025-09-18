package handler

import (
	"log"
	"net/http"
	"strings"

	models "github.com/htrandev/metrics/internal/model"
)

type MetricStorage interface {
	Get(string) (models.Metric, error)
	GetAll() ([]models.Metric, error)
	Store(m *models.Metric) error
}

type MetricHandler struct {
	storage MetricStorage
}

func NewMetricsHandler(s MetricStorage) *MetricHandler {
	return &MetricHandler{
		storage: s,
	}
}

func (h *MetricHandler) Get(rw http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")

	mt := models.ParseMetricType(metricType)
	if mt == models.TypeUnknown {
		log.Print("got unknown metric type\n\r")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.storage.Get(metricName)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(metric.Value.String()))
}

func (h *MetricHandler) GetAll(rw http.ResponseWriter, r *http.Request) {
	metrics, err := h.storage.GetAll()
	if err != nil {
		log.Printf("get all metrics: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	var builder strings.Builder
	for _, metric := range metrics {
		builder.Grow(len(metric.Name) + len(metric.Value.String()) + 2)
		builder.WriteString(metric.Name)
		builder.WriteString(": ")
		builder.WriteString(metric.Value.String())
		builder.WriteString("\r")
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(builder.String()))
}

func (h *MetricHandler) Update(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("expected %s, but got %s\n\r", http.MethodPost, r.Method)
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	if metricName == "" {
		log.Print("got empty metric name\n\r")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	mt := models.ParseMetricType(metricType)
	if mt == models.TypeUnknown {
		log.Print("got unknown metric type\n\r")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	metric := &models.Metric{
		Name: metricName,
		Value: models.MetricValue{
			Type: mt,
		},
	}
	if err := metric.SetValue(metricValue); err != nil {
		log.Printf("set value [%v] for [%s]: %v\n\r", metricValue, metric.Name, err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.storage.Store(metric); err != nil {
		log.Printf("store error: %s\n\r", err.Error())
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("successfully store [%+v]", metric)

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
}
