package registry

import "sync"

type Calculator interface{}

type Registry struct {
	mu    sync.RWMutex
	items map[string]Calculator
}

func New() *Registry {
	return &Registry{items: make(map[string]Calculator)}
}

func (r *Registry) Register(name string, c Calculator) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[name]; ok {
		return ErrMetricAlreadyRegistered
	}
	r.items[name] = c
	return nil
}

func (r *Registry) Get(name string) (Calculator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.items[name]
	if !ok {
		return nil, ErrMetricNotFound
	}
	return c, nil
}
