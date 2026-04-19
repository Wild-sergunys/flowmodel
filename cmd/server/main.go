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
	paramRepo := repository.NewParameterRepo(db)

	// Хэндлеры
	authHandler := handler.NewAuthHandler(userRepo, jwtKey)
	materialHandler := handler.NewMaterialHandler(materialRepo)
	adminHandler := handler.NewAdminHandler(materialRepo)
	userHandler := handler.NewUserHandler(userRepo)
	paramHandler := handler.NewParameterHandler(paramRepo)

	// Middleware
	authMiddleware := middleware.AuthMiddleware([]byte(jwtKey))
	adminMiddleware := middleware.RoleMiddleware("admin")

	// Роутер
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.Handle("GET /api/auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /api/materials", authMiddleware(http.HandlerFunc(materialHandler.GetAll)))

	// Админский роут
	mux.Handle("GET /api/admin/materials", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.GetAllMaterials))))
	mux.Handle("GET /api/admin/materials/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.GetMaterialByID))))
	mux.Handle("POST /api/admin/materials", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.CreateMaterial))))
	mux.Handle("PUT /api/admin/materials/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.UpdateMaterial))))
	mux.Handle("DELETE /api/admin/materials/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.DeleteMaterial))))
	mux.Handle("GET /api/admin/parameters", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.GetAll))))
	mux.Handle("GET /api/admin/parameters/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.GetByID))))
	mux.Handle("POST /api/admin/parameters", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.Create))))
	mux.Handle("PUT /api/admin/parameters/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.Update))))
	mux.Handle("DELETE /api/admin/parameters/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.Delete))))
	mux.Handle("POST /api/admin/users", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.Create))))
	mux.Handle("GET /api/admin/users", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.GetAll))))
	mux.Handle("GET /api/admin/users/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.GetByID))))
	mux.Handle("PUT /api/admin/users/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.Update))))
	mux.Handle("DELETE /api/admin/users/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.Delete))))

	log.Printf("Сервер запущен на http://localhost:%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
