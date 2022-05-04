package usecase

import (
	"context"
	service2 "github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
)

type creator struct {
	finalizer service2.Finalizer
	resource  service2.ResourceManager
}

// NewCreator instantiates creation use case states
func NewCreator(f service2.Finalizer, r service2.ResourceManager) service2.ConciliatorHandler {
	return &creator{
		finalizer: f,
		resource:  r,
	}
}

// Empty Status handler
func (c *creator) Empty(ctx context.Context, ps *v1alpha1.PrometheusServer) (newStatus string, err error) {
	defer emptyProcessed.Inc()

	if err := c.finalizer.Ensure(ctx, ps); err != nil {
		return ps.Status.Phase, err
	}
	return v1alpha1.Initializing, nil
}

// Initializing Status handler
func (c *creator) Initializing(ctx context.Context, ps *v1alpha1.PrometheusServer) (newStatus string, err error) {
	defer initializingProcessed.Inc()

	if err := c.resource.CreateAll(ctx, ps); err != nil {
		return ps.Status.Phase, err
	}

	return v1alpha1.WaitingCreation, nil
}

// WaitingCreation Status handler
func (c *creator) WaitingCreation(ctx context.Context, ps *v1alpha1.PrometheusServer) (newStatus string, err error) {
	defer waitingCreationProcessed.Inc()

	ok, err := c.resource.AllCreated()
	if err != nil {
		return ps.Status.Phase, err
	}
	if !ok {
		return ps.Status.Phase, nil
	}

	return v1alpha1.Running, nil
}

// Handlers return creation status handlers
func (c *creator) Handlers() map[string]service2.StateHandler {
	return map[string]service2.StateHandler{
		v1alpha1.Empty:           c.Empty,
		v1alpha1.Initializing:    c.Initializing,
		v1alpha1.WaitingCreation: c.WaitingCreation,
	}
}
