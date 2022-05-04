package service

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// StateHandler behaves as FSM handler, executes required actions and return new state
type StateHandler func(ctx context.Context, ps *v1alpha1.PrometheusServer) (newStatus string, err error)

// ConciliatorHandler wraps conciliation use cases
type ConciliatorHandler interface {
	Handlers() map[string]StateHandler
}

type conciliation struct {
	state map[string]StateHandler
}

// NewConciliator instantiates conciliator
func NewConciliator() *conciliation {
	return &conciliation{
		state: map[string]StateHandler{},
	}
}

// Register conciliator handler
func (c *conciliation) Register(o ConciliatorHandler) {
	for s, handler := range o.Handlers() {
		log.Debugf("registering %s handler", s)
		c.state[s] = handler
	}
}

// Conciliate will apply registered state handler
func (c *conciliation) Conciliate(ctx context.Context, ps *v1alpha1.PrometheusServer) (string, error) {
	h, ok := c.state[ps.Status.Phase]
	if !ok {
		conciliationProcessedErrors.Inc()
		return ps.Status.Phase, fmt.Errorf("no handler registered on %s phase", ps.Status.Phase)
	}

	defer conciliationProcessed.Inc()

	newState, err := h(ctx, ps)
	if err != nil {
		conciliationProcessedErrors.Inc()
		return ps.Status.Phase, fmt.Errorf("error hhandling %s phase, error %w", ps.Status.Phase, err)
	}

	return newState, nil
}
