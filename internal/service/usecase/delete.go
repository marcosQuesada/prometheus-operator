package usecase

import (
	"context"
	"fmt"
	service2 "github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
)

type deleter struct {
	finalizer service2.Finalizer
	resource  service2.ResourceManager
}

// NewDeleter instantiates deletion use case status handlers
func NewDeleter(f service2.Finalizer, r service2.ResourceManager) service2.ConciliatorHandler {
	return &deleter{
		finalizer: f,
		resource:  r,
	}
}

// Terminating Status handler
func (d *deleter) Terminating(ctx context.Context, ps *v1alpha1.PrometheusServer) (string, error) {
	defer terminatingProcessed.Inc()

	if err := d.resource.DeleteAll(ctx, ps); err != nil {
		return ps.Status.Phase, err
	}
	if !service2.HasFinalizer(ps) {
		return v1alpha1.Terminated, nil
	}
	if err := d.finalizer.Remove(ctx, ps); err != nil {
		return ps.Status.Phase, fmt.Errorf("unable to removing finalizer, error %w", err)
	}
	return v1alpha1.Terminated, nil
}

// Handlers return creation status handlers
func (d *deleter) Handlers() map[string]service2.StateHandler {
	return map[string]service2.StateHandler{
		v1alpha1.Terminating: d.Terminating,
	}
}
