package store

import (
	"context"
	"fmt"
)

func IsNoSeedFound(err error) bool {
	_, ok := err.(*noSeedFoundError)
	return ok
}

type noSeedFoundError struct {
	name string
}

func (e *noSeedFoundError) Error() string {
	return fmt.Sprintf("no matching seed with %s could be found", e.name)
}

type Store interface {
	Seed(ctx context.Context, name string) (string, error)
	PutSeed(ctx context.Context, name, target string) error
	DeleteSeed(ctx context.Context, name string) error
}
