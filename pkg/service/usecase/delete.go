package usecase

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/service"
)

type deleter struct {
	finalizer service.Finalizer
	resource  service.ResourceManager
}

// NewDeleter instantiates deletion use case status handlers
func NewDeleter(f service.Finalizer, r service.ResourceManager) service.ConciliatorHandler {
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
	if !ps.HasFinalizer(v1alpha1.Name) {
		return v1alpha1.Terminated, nil
	}
	if err := d.finalizer.Remove(ctx, ps); err != nil {
		return ps.Status.Phase, fmt.Errorf("unable to removing finalizer, error %v", err)
	}
	return v1alpha1.Terminated, nil
}

// Handlers return creation status handlers
func (d *deleter) Handlers() map[string]service.StateHandler {
	return map[string]service.StateHandler{
		v1alpha1.Terminating: d.Terminating,
	}
}
