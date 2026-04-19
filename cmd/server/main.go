package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"flowmodel/internal/config"
	"flowmodel/internal/database"
	"flowmodel/internal/handler"
	"flowmodel/internal/middleware"
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

	// JWT подтягиваем из .env или значение по умодчанию
	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		jwtKey = "your-secret-key-change-in-prod"
	}

	// Репозитории
	userRepo := repository.NewUserRepo(db)
	materialRepo := repository.NewMaterialRepo(db)

	// Хэндлеры
	authHandler := handler.NewAuthHandler(userRepo, jwtKey)
	materialHandler := handler.NewMaterialHandler(materialRepo)
	adminHandler := handler.NewAdminHandler(materialRepo)

	// Middleware
	authMiddleware := middleware.AuthMiddleware([]byte(jwtKey))
	adminMiddleware := middleware.RoleMiddleware("admin")

	// Роутер
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.Handle("GET /api/materials", authMiddleware(http.HandlerFunc(materialHandler.GetAll)))

	// Админский роут
	mux.Handle("GET /api/admin/materials", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.GetAllMaterials))))

	log.Printf("Сервер запущен на http://localhost:%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
