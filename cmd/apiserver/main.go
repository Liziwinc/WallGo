package main

import (
	"WallGo/internal/handler"
	"WallGo/internal/repository"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)


func main() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "",
		DB:       0,
	})

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "secret"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "pastebin_db"
	}

	ctx := context.Background()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	
	pgRepo, err := repository.NewPostgresql(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	cachedRepo := repository.NewCachedPostRepository(pgRepo, rdb, 5*time.Minute)
	h := handler.NewHandler(cachedRepo)
	r := chi.NewRouter()
	r.Post("/posts", h.CreatePost)
	r.Get("/posts/{hash}", h.GetPost)
	r.Get("/posts", h.GetPostsList)
	fileServer := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fileServer)
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}

