package store

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	rcli   *redis.Client
	prefix string
}

func NewRedis(rcli *redis.Client, prefix string) Store {
	return &Redis{
		rcli:   rcli,
		prefix: prefix,
	}
}

func (r *Redis) Seed(ctx context.Context, name string) (string, error) {
	tgt, err := r.rcli.Get(ctx, r.prefix+":"+name).Result()
	if errors.Is(err, redis.Nil) {
		return "", &noSeedFoundError{name: name}
	}

	return tgt, err
}

func (r *Redis) PutSeed(ctx context.Context, name, target string) error {
	_, err := r.rcli.Set(ctx, r.prefix+":"+name, target, 0).Result()
	return err
}

func (r *Redis) DeleteSeed(ctx context.Context, name string) error {
	_, err := r.rcli.Del(ctx, r.prefix+":"+name).Result()
	return err
}
