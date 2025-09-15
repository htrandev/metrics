package handler

import (
	"log"
	"net/http"

	models "github.com/htrandev/metrics/internal/model"
)

type MetricStorage interface {
	Store(m *models.Metric) error
}

type UpdateHandler struct {
	storage MetricStorage
}

func NewUpdateHandler(s MetricStorage) *UpdateHandler {
	return &UpdateHandler{
		storage: s,
	}
}

func (u *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("expected %s, but got %s\n\r", http.MethodPost, r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	log.Println(r.URL.Path)

	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	log.Printf("metric type: %s,\n\rmetric name: %s,\n\rmetric value: %s\n\r",
		metricType,
		metricName,
		metricValue,
	)

	if metricName == "" {
		log.Print("got empty metric name\n\r")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	m := models.ParseMetricType(metricType)
	if m == models.TypeUnknown {
		log.Print("got unknown metric type\n\r")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric := &models.Metric{
		Name: metricName,
		Type: m,
	}
	if err := metric.SetValue(metricValue); err != nil {
		log.Printf("set value: %v\n\r", metricValue)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := u.storage.Store(metric); err != nil {
		log.Printf("store error: %s\n\r", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
