package database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber"
	"os"
	"time"
)

type RedisClient struct {
	addr     string
	password string
	DB       int
	Client   *redis.Client
	SetValue interface{}
	TTL      time.Duration
}

type RedisClientError struct {
	Error   error
	Message string
	Status  int
}

var Ctx = context.Background()

func (rc *RedisClient) CreateRedisClient() {
	rc.Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_PASS"),
		DB:       rc.DB,
	})
}

func (rc *RedisClient) Set(key string) *RedisClientError {
	err := rc.Client.Set(Ctx, key, rc.SetValue, rc.TTL).Err()

	if err != nil {
		return &RedisClientError{
			Error:   err,
			Message: fmt.Sprintf("Cannot set key: %s", key),
			Status:  fiber.StatusInternalServerError,
		}
	}

	return nil
}

func (rc *RedisClient) Get(key string, shouldSet bool) (string, *RedisClientError) {
	value, err := rc.Client.Get(Ctx, key).Result()

	if err == redis.Nil {
		if shouldSet {
			errSet := rc.Set(key)
			if errSet != nil {
				return "", errSet
			}
		}

		return "", &RedisClientError{
			Error:   err,
			Message: fmt.Sprintf("Key not found - key: %s ", key),
			Status:  fiber.StatusNotFound,
		}
	} else if err != nil {
		return "", &RedisClientError{
			Error:   err,
			Message: fmt.Sprintf("cannot connect to db"),
			Status:  fiber.StatusInternalServerError,
		}
	}

	return value, nil
}

func (rc *RedisClient) Increment(key string) *redis.IntCmd {
	return rc.Client.Incr(Ctx, key)
}

func (rc *RedisClient) Decrement(key string) *redis.IntCmd {
	return rc.Client.Decr(Ctx, key)
}
