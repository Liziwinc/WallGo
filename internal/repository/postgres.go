package repository

import (
	"WallGo/internal/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateHash = errors.New("duplicate hash")

type Postgresql struct {
	db *pgxpool.Pool
}

func NewPostgresql(ctx context.Context, dsn string) (*Postgresql, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return &Postgresql{db: pool}, nil
}

func (r *Postgresql) Create(ctx context.Context, post *model.Post) error {
	query := ` INSERT INTO post (hash, content, expires_at)
	VALUES ($1, $2, $3)
	RETURNING id, created_at;`
	err := r.db.QueryRow(ctx, query, post.Hash, post.Content, post.ExpiresAt).Scan(&post.Id, &post.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrDuplicateHash
			}
		}
		return err
	}
	return nil
}
func (r *Postgresql) GetByHash(ctx context.Context, hash string) (*model.Post, error) {
	var post model.Post
	query := ` SELECT id, hash, content, created_at, expires_at FROM post 
										WHERE hash = $1 AND (expires_at > NOW() OR expires_at IS NULL);`
	err := r.db.QueryRow(ctx, query, hash).Scan(&post.Id, &post.Hash, &post.Content, &post.CreatedAt, &post.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *Postgresql) GetList(ctx context.Context, limit int, offset int) ([]*model.Post, error) {
	var posts []*model.Post
	query := `SELECT id, hash, content, created_at, expires_at FROM post 
								  WHERE expires_at > NOW() OR expires_at IS NULL
								  ORDER BY created_at DESC 
								  LIMIT $1 OFFSET $2;`
	data, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer data.Close()
	for data.Next() {
		var post model.Post
		err = data.Scan(&post.Id, &post.Hash, &post.Content, &post.CreatedAt, &post.ExpiresAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	if err = data.Err(); err != nil {
		return nil, err
	}
	return posts, err
}
