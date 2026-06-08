package handler

import (
	"WallGo/internal/model"
	"WallGo/internal/repository"
	"WallGo/pkg/base62"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	GetByHash(ctx context.Context, hash string) (*model.Post, error)
	GetList(ctx context.Context, limit, offset int) ([]*model.Post, error)
}

type Handler struct {
	pr PostRepository
}

func NewHandler(repo PostRepository) *Handler {
	return &Handler{pr: repo}
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	type createPostRequest struct {
		Content   string     `json:"content"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	var req createPostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	post := &model.Post{
		Content:   req.Content,
		ExpiresAt: req.ExpiresAt,
	}
	for i := 0; i < 3; i++ {
		post.Hash = base62.GenerateHash()

		err = h.pr.Create(r.Context(), post)
		if err == nil {
			break
		}

		if errors.Is(err, repository.ErrDuplicateHash) {
			log.Println("Collision detected, retrying...")
			continue
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return

	}
	if errors.Is(err, repository.ErrDuplicateHash) {
		http.Error(w, "could not generate unique hash", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		return
	}
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	post, err := h.pr.GetByHash(r.Context(), hash)
	if err != nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		return
	}
}

func (h *Handler) GetPostsList(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	limitStr := vals.Get("limit")
	offsetStr := vals.Get("offset")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limitStr == "" || limit < 0 || limit > 50{
		limit = 10
		
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offsetStr == "" || offset < 0 {
		offset = 0
	}

	post, err := h.pr.GetList(r.Context(), limit,offset)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		return
	}
}
