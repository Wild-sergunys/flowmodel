package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	"flowmodel/internal/repository"
)

func getDSN() string {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "root"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "flowmodel"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, dbname)
}

func main() {
	_ = godotenv.Load()
	dsn := getDSN()
	log.Println("Подключение к MySQL...")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к MySQL:", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Не удалось подключиться к MySQL:", err)
	}
	log.Println("Подключено к MySQL")

	repo := repository.NewMaterialRepo(db)
	materials, err := repo.FindAll(ctx) // ← исправлено
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Найдено %d материалов", len(materials))
	for _, m := range materials {
		log.Printf("ID: %d, Name: %s", m.ID, m.Name)
	}
}
