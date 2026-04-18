package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"flowmodel/internal/config"
	"flowmodel/internal/database"
	"flowmodel/internal/handler"
	"flowmodel/internal/repository"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	log.Println("Подключение к MySQL...")

	// Миграции
	if err := database.RunMigrations(cfg.DSN()); err != nil {
		log.Fatal("Ошибка миграций:", err)
	}

	// Подключение к БД
	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	// Запуск сервера
	repo := repository.NewMaterialRepo(db)
	materialHandler := handler.NewMaterialHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/materials", materialHandler.GetAll)

	log.Printf("Сервер запущен на http://localhost:%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
