package json_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/NoobyTheTurtle/metrics/internal/handler/json"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/NoobyTheTurtle/metrics/internal/storage/memory"
)

// ExampleHandler_UpdateHandler демонстрирует обновление gauge метрики через JSON API.
func ExampleHandler_UpdateHandler() {
	// Настройка хранилища и обработчика
	memStorage := memory.NewMemoryStorage()
	storage := adapter.NewStorage(memStorage)
	handler := json.NewHandler(storage)

	// Создание тестового сервера
	server := httptest.NewServer(handler.UpdateHandler())
	defer server.Close()

	// Обновление gauge метрики
	jsonPayload := `{"id": "cpu_usage", "type": "gauge", "value": 85.5}`
	resp, err := http.Post(server.URL, "application/json", bytes.NewBufferString(jsonPayload))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// ExampleHandler_UpdateHandler_counter демонстрирует обновление counter метрики через JSON API.
func ExampleHandler_UpdateHandler_counter() {
	// Настройка хранилища и обработчика
	memStorage := memory.NewMemoryStorage()
	storage := adapter.NewStorage(memStorage)
	handler := json.NewHandler(storage)

	// Создание тестового сервера
	server := httptest.NewServer(handler.UpdateHandler())
	defer server.Close()

	// Обновление counter метрики
	jsonPayload := `{"id": "requests_total", "type": "counter", "delta": 1}`
	resp, err := http.Post(server.URL, "application/json", bytes.NewBufferString(jsonPayload))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// ExampleHandler_ValueHandler демонстрирует получение значения метрики через JSON API.
func ExampleHandler_ValueHandler() {
	// Настройка хранилища и обработчика
	memStorage := memory.NewMemoryStorage()
	storage := adapter.NewStorage(memStorage)
	handler := json.NewHandler(storage)

	// Предварительное заполнение gauge метрикой
	storage.UpdateGauge(context.Background(), "memory_usage", 512.0)

	// Создание тестового сервера
	server := httptest.NewServer(handler.ValueHandler())
	defer server.Close()

	// Запрос значения метрики
	jsonPayload := `{"id": "memory_usage", "type": "gauge"}`
	resp, err := http.Post(server.URL, "application/json", bytes.NewBufferString(jsonPayload))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// ExampleHandler_UpdatesHandler демонстрирует пакетное обновление метрик через JSON API.
func ExampleHandler_UpdatesHandler() {
	// Настройка хранилища и обработчика
	memStorage := memory.NewMemoryStorage()
	storage := adapter.NewStorage(memStorage)
	handler := json.NewHandler(storage)

	// Создание тестового сервера
	server := httptest.NewServer(handler.UpdatesHandler())
	defer server.Close()

	// Пакетное обновление нескольких метрик
	jsonPayload := `[
		{"id": "cpu_usage", "type": "gauge", "value": 75.2},
		{"id": "memory_usage", "type": "gauge", "value": 1024.0},
		{"id": "requests_total", "type": "counter", "delta": 5}
	]`
	resp, err := http.Post(server.URL, "application/json", bytes.NewBufferString(jsonPayload))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}
