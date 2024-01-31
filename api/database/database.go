package database

import (
	"context"
	"os"
	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

func CreateClient(dbNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis://"+os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_PASS"),
		DB: dbNo,
	})
	return rdb
}