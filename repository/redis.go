package repository

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	rcli   *redis.Client
	prefix string
}

var _ Seed = (*Redis)(nil)

func NewRedis(rcli *redis.Client, prefix string) *Redis {
	return &Redis{
		rcli:   rcli,
		prefix: prefix,
	}
}

func (r *Redis) Seed(ctx context.Context, name string) (string, error) {
	tgt, err := r.rcli.Get(ctx, r.prefix+":"+name).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNoSeedFound
	}
	return tgt, nil
}

func (r *Redis) PutSeed(ctx context.Context, name, target string) error {
	_, err := r.rcli.Set(ctx, r.prefix+":"+name, target, 0).Result()
	return err
}
