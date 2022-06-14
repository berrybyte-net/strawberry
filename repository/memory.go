package repository

import (
	"context"
)

type Memory struct {
	seeds map[string]string
}

var _ Seed = (*Memory)(nil)

func NewMemory() *Memory {
	return &Memory{
		seeds: make(map[string]string),
	}
}

func (m *Memory) Seed(ctx context.Context, name string) (string, error) {
	if tgt, ok := m.seeds[name]; ok {
		return tgt, nil
	}

	return "", ErrNoSeedFound
}

func (m *Memory) PutSeed(ctx context.Context, name, target string) error {
	if _, ok := m.seeds[name]; ok {
		return nil
	}

	m.seeds[name] = target
	return nil
}
