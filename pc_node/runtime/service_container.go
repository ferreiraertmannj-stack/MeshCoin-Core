package runtime

import (
	"fmt"
	"sync"
)

type ServiceContainer struct {
	mu        sync.RWMutex
	providers map[string]interface{}
}

func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		providers: make(map[string]interface{}),
	}
}

func (c *ServiceContainer) RegisterProvider(name string, provider interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	c.providers[name] = provider
	return nil
}

func (c *ServiceContainer) Resolve(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	provider, exists := c.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}
