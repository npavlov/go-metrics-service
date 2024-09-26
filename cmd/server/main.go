package main

import (
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/handler"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"log"
	"net/http"
)

func main() {
	var memStorage storage.Repository = storage.NewMemStorage()

	updateHandler := handler.GetUpdateHandler(memStorage)

	// Маршрутизация
	http.HandleFunc("/update/", updateHandler)

	// Запуск сервера
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
