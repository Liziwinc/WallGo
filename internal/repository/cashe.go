package repository

import (
	"WallGo/internal/model"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedPostRepository struct {
	postgres *Postgresql
	redis    *redis.Client
	ttl      time.Duration
}

func NewCachedPostRepository(pg *Postgresql, rdb *redis.Client, ttl time.Duration) *CachedPostRepository {
	return &CachedPostRepository{
		postgres: pg,
		redis:    rdb,
		ttl:      ttl,
	}
}

func (c *CachedPostRepository) GetByHash(ctx context.Context, hash string) (*model.Post, error) {
	key := "post:" + hash
	val, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var post model.Post
		err := json.Unmarshal([]byte(val), &post)
		if err != nil {
			return nil, err
		}
		return &post, nil
	}
	dbpost, err := c.postgres.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(dbpost)
	redisTTL := c.ttl
	if dbpost.ExpiresAt != nil {
		remains := time.Until(*dbpost.ExpiresAt)
		if remains < redisTTL {
			redisTTL = remains
		}
	}
	if redisTTL > 0 {
		c.redis.Set(ctx, key, data, redisTTL)
	}
	return dbpost, nil
}

func (c *CachedPostRepository) Create(ctx context.Context, post *model.Post) error {
	err := c.postgres.Create(ctx, post)
	if err != nil {
		return err
	}
	return nil
}

func (c *CachedPostRepository) GetList(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	return c.postgres.GetList(ctx, limit, offset)
}
