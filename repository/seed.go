package repository

import (
	"context"

	"github.com/pkg/errors"
)

var ErrNoSeedFound = errors.New("no matching seed could be found")

type Seed interface {
	Seed(ctx context.Context, name string) (string, error)
	PutSeed(ctx context.Context, name, target string) error
}
