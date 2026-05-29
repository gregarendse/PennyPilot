package sync

import (
	"fmt"
)

// Registry manages the set of available bank connectors.
type Registry struct {
	connectors map[string]BankConnector
}

func NewRegistry() *Registry {
	return &Registry{
		connectors: make(map[string]BankConnector),
	}
}

func (r *Registry) Register(c BankConnector) {
	r.connectors[c.Name()] = c
}

func (r *Registry) Get(name string) (BankConnector, error) {
	c, ok := r.connectors[name]
	if !ok {
		return nil, fmt.Errorf("connector %q not found", name)
	}
	return c, nil
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.connectors))
	for name := range r.connectors {
		names = append(names, name)
	}
	return names
}
