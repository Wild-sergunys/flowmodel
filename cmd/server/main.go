package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"flowmodel/internal/config"
	"flowmodel/internal/database"
	"flowmodel/internal/handler"
	"flowmodel/internal/middleware"
	"flowmodel/internal/repository"
	"flowmodel/internal/service"
	"flowmodel/web"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	log.Println("Подключение к MySQL...")

	if err := database.RunMigrations(cfg.DSN()); err != nil {
		log.Fatal("Ошибка миграций:", err)
	}

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		jwtKey = "your-secret-key-change-in-prod"
	}

	// Репозитории
	userRepo := repository.NewUserRepo(db)
	materialRepo := repository.NewMaterialRepo(db)
	paramRepo := repository.NewParameterRepo(db)
	materialParamRepo := repository.NewMaterialParameterRepo(db)
	calcRepo := repository.NewCalculationRepo(db)

	// Сервисы
	authService := service.NewAuthService(userRepo, jwtKey)
	calcService := service.NewCalculationService(materialParamRepo, materialRepo, calcRepo)

	// Хэндлеры
	authHandler := handler.NewAuthHandler(authService)
	materialHandler := handler.NewMaterialHandler(materialRepo)
	materialParamHandler := handler.NewMaterialParameterHandler(materialParamRepo)
	adminHandler := handler.NewAdminHandler(materialRepo)
	userHandler := handler.NewUserHandler(userRepo)
	paramHandler := handler.NewParameterHandler(paramRepo)
	calcHandler := handler.NewCalculationHandler(calcService)
	resultsHandler := handler.NewResultsHandler(calcRepo)
	webHandler, err := web.NewHandler()
	if err != nil {
		log.Fatal("Ошибка инициализации frontend:", err)
	}

	// Rate limiter для логина (настройки из cfg)
	loginLimiter := middleware.NewRateLimiter(
		cfg.LoginMaxAttempts,
		time.Duration(cfg.LoginWindowMin)*time.Minute,
		time.Duration(cfg.LoginBlockMin)*time.Minute,
	)
	handler.SetLoginRateLimiter(loginLimiter)
	loginRateLimitMiddleware := middleware.LoginRateLimitMiddleware(loginLimiter)

	// Middleware
	authMiddleware := middleware.AuthMiddleware([]byte(jwtKey))
	adminMiddleware := middleware.RoleMiddleware("admin")

	// Роутер
	mux := http.NewServeMux()

	// Frontend
	mux.Handle("GET /static/", webHandler.Static())
	mux.HandleFunc("GET /login", webHandler.Login)
	mux.HandleFunc("GET /admin", webHandler.Admin)
	mux.HandleFunc("GET /", webHandler.Home)
	mux.HandleFunc("GET /cabinet", webHandler.Cabinet)

	// Auth
	mux.Handle("POST /api/auth/login", loginRateLimitMiddleware(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.Handle("GET /api/auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))

	// Materials (public)
	mux.Handle("GET /api/materials", authMiddleware(http.HandlerFunc(materialHandler.GetAll)))
	mux.Handle("GET /api/admin/materials/{id}/parameters", authMiddleware(adminMiddleware(http.HandlerFunc(materialParamHandler.ListParameters))))
	mux.Handle("PUT /api/admin/materials/{id}/parameters", authMiddleware(adminMiddleware(http.HandlerFunc(materialParamHandler.UpdateParameters))))

	// Validation (public)
	mux.HandleFunc("POST /api/validate", calcHandler.Validate)

	// Calculation
	mux.Handle("POST /api/calculate", authMiddleware(http.HandlerFunc(calcHandler.Calculate)))

	// Admin Materials
	mux.Handle("GET /api/admin/materials", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.GetAllMaterials))))
	mux.Handle("GET /api/admin/materials/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.GetMaterialByID))))
	mux.Handle("POST /api/admin/materials", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.CreateMaterial))))
	mux.Handle("PUT /api/admin/materials/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.UpdateMaterial))))
	mux.Handle("DELETE /api/admin/materials/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.DeleteMaterial))))

	// Admin Parameters
	mux.Handle("GET /api/admin/parameters", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.GetAll))))
	mux.Handle("GET /api/admin/parameters/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.GetByID))))
	mux.Handle("POST /api/admin/parameters", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.Create))))
	mux.Handle("PUT /api/admin/parameters/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.Update))))
	mux.Handle("DELETE /api/admin/parameters/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(paramHandler.Delete))))

	// Admin Users
	mux.Handle("POST /api/admin/users", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.Create))))
	mux.Handle("GET /api/admin/users", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.GetAll))))
	mux.Handle("GET /api/admin/users/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.GetByID))))
	mux.Handle("PUT /api/admin/users/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.Update))))
	mux.Handle("DELETE /api/admin/users/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(userHandler.Delete))))

	// Results
	mux.Handle("GET /api/results", authMiddleware(http.HandlerFunc(resultsHandler.GetAll)))
	mux.Handle("GET /api/results/{id}", authMiddleware(http.HandlerFunc(resultsHandler.GetByID)))
	mux.Handle("GET /api/results/{id}/report", authMiddleware(http.HandlerFunc(resultsHandler.GetReport)))
	mux.Handle("GET /api/results/{id}/download", authMiddleware(http.HandlerFunc(resultsHandler.Download)))

	// Запуск сервера с graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Сервер запущен на http://localhost:%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	<-stop
	log.Println("Получен сигнал остановки (SIGTERM/SIGINT)")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Ожидание завершения активных запросов (до 10с)...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Сервер завершён принудительно:", err)
	} else {
		log.Println("Все запросы завершены корректно")
	}
}
