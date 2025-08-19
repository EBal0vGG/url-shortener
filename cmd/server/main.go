package main

import (
	"log"
	"net/http"
	"url-shortener/internal/httpapi"
	"url-shortener/internal/storage"
)

func main() {


	// Подключаемся к базе через переменные окружения
	dbConn, err := storage.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	// Оборачиваем *sql.DB в нашу структуру Postgres
	db := storage.NewPostgres(dbConn)

	// Создаём маршрутизатор, передаём Postgres
	mux := httpapi.NewRouter(db)

	log.Println("Server is running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
