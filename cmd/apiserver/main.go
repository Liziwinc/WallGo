package main

import (
	"WallGo/internal/handler"
	"WallGo/internal/repository"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	dsn := "postgres://postgres:secret@localhost:5432/pastebin_db?sslmode=disable"
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
