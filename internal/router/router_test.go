package router

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/htrandev/metrics/internal/audit"
	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/repository/postgres"
	"github.com/htrandev/metrics/internal/service/metrics"
	"github.com/htrandev/metrics/pkg/logger"
	"github.com/mailru/easyjson"
)

// Example - пример работы с эндпоинтами.
func Example() {
	databaseDsn := ""
	contentType := "application/json"

	// инициализируем логгер
	logger, _ := logger.NewZapLogger("debug")

	// подключаемся к бд
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// инициализируем хранилище
	storage := postgres.New(db, 3)

	// инициализируем сервис
	service := metrics.NewService(&metrics.ServiсeOptions{
		Logger:  logger,
		Storage: storage,
	})

	// инициализируем аудитора
	auditor := audit.NewAuditor()

	// инициализируем хэндлер
	handler := handler.NewMetricsHandler(
		logger,
		service,
		auditor,
	)

	// инициализируем роутер
	router := New(RouterOptions{
		Signature: "",
		Subnet:    nil,
		Key:       nil,
		Logger:    logger,
		Handler:   handler,
	})

	// инициализируем сервис
	srv := http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("can't start server: %v", err)
		}
	}()

	// Даем серверу время запуститься
	time.Sleep(100 * time.Millisecond)

	/* 1. Получение всех метрик */
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		fmt.Printf("Error /: %v\n", err)
	} else {
		fmt.Printf("/ : %d\n", resp.ContentLength)
		_ = resp.Body.Close()
	}

	/* 2. Получение значения метрики */
	resp, err = http.Get("http://localhost:8080/value/gauge/Alloc")
	if err != nil {
		fmt.Printf("Error /value: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("/value: %s\n", string(body))
		_ = resp.Body.Close()
	}

	/* 3. Обновление метрики через путь */
	resp, err = http.Post("http://localhost:8080/update/gauge/cpu_usage/3.14", "", nil)
	if err != nil {
		fmt.Printf("Error /update path: %v\n", err)
	} else {
		fmt.Printf("/update path: %s\n", resp.Status)
		_ = resp.Body.Close()
	}

	body := bytes.NewBuffer([]byte(`{"id":"gauge","type":"gauge","value":0.1}`))

	/* 4. Обновление метрики в формате JSON */
	resp, err = http.Post("http://localhost:8080/update/", contentType, body)
	if err != nil {
		fmt.Printf("Error /update body: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("/update body: %s\n", string(body))
		_ = resp.Body.Close()
	}

	/* 5. Получение метрики в формате JSON */
	resp, err = http.Post("http://localhost:8080/value/", contentType, body)
	if err != nil {
		fmt.Printf("Error /value body: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("/value body: %s\n", string(body))
		_ = resp.Body.Close()
	}

	/* 6. Проверка доступности БД */
	resp, err = http.Get("http://localhost:8080/ping")
	if err != nil {
		fmt.Printf("Error /ping: %v\n", err)
	} else {
		fmt.Printf("/ping: %s\n", resp.Status)
		_ = resp.Body.Close()
	}

	/* 7. Обновление нескольких метрик за раз */
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
	b, _ := easyjson.Marshal(metrics)

	resp, err = http.Post("http://localhost:8080/updates/", contentType, bytes.NewBuffer(b))
	if err != nil {
		fmt.Printf("Error /updates: %v\n", err)
	} else {
		fmt.Printf("/updates: %s\n", resp.Status)
		_ = resp.Body.Close()
	}
}
