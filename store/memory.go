package store

import (
	"context"
)

type Memory struct {
	seeds map[string]string
}

func NewMemory() Store {
	return &Memory{
		seeds: make(map[string]string),
	}
}

func (m *Memory) Seed(ctx context.Context, name string) (string, error) {
	if tgt, ok := m.seeds[name]; ok {
		return tgt, nil
	}

	return "", &noSeedFoundError{name: name}
}

func (m *Memory) PutSeed(ctx context.Context, name, target string) error {
	if _, ok := m.seeds[name]; ok {
		return nil
	}

	m.seeds[name] = target
	return nil
}

func (m *Memory) DeleteSeed(ctx context.Context, name string) error {
	delete(m.seeds, name)
	return nil
}
