package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
)

type generationCache struct {
	generation map[string]int64 // track CRD generation to allow rollout on new versions
	mutex      sync.RWMutex
}

// NewGenerationCache instantiates a CRD generation cache
func NewGenerationCache() Cache {
	return &generationCache{generation: map[string]int64{}}
}

// Get returns persisted CRD generation, on not found returns 0
func (o *generationCache) Get(namespace, name string) int64 {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	v, ok := o.generation[fmt.Sprintf("%s/%s", namespace, name)]
	if !ok {
		return 0
	}
	return v
}

// Set adds resource generation
func (o *generationCache) Set(namespace, name string, v int64) {
	log.Debugf("Setting to Registry %s/%s on generation %d", namespace, name, v)
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.generation[fmt.Sprintf("%s/%s", namespace, name)] = v
}

// Remove resource generation
func (o *generationCache) Remove(namespace, name string) {
	log.Debugf("Removing from registry %s/%s", namespace, name)
	o.mutex.Lock()
	defer o.mutex.Unlock()

	delete(o.generation, fmt.Sprintf("%s/%s", namespace, name))
}
