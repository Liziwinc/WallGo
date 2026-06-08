package model

import "time"

type Post struct{
	Id int64 `json:"id"`
	Hash string `json:"hash"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}
