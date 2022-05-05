package usecase

import (
	"context"
	"fmt"

	"github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

type deleter struct {
	finalizer service.Finalizer
	resource  service.ResourceManager
	recorder  record.EventRecorder
}

// NewDeleter instantiates deletion use case status handlers
func NewDeleter(f service.Finalizer, r service.ResourceManager, e record.EventRecorder) service.ConciliatorHandler {
	return &deleter{
		finalizer: f,
		resource:  r,
		recorder:  e,
	}
}

// Terminating Status handler
func (d *deleter) Terminating(ctx context.Context, ps *v1alpha1.PrometheusServer) (string, error) {
	defer terminatingProcessed.Inc()

	if err := d.resource.DeleteAll(ctx, ps); err != nil {
		d.recorder.Eventf(ps, v1.EventTypeNormal, "DeleteAllError", "Prometheus Server Namespace %s Name %s delete error %s", ps.Namespace, ps.Name, err.Error())

		return ps.Status.Phase, err
	}
	if !service.HasFinalizer(ps) {
		return v1alpha1.Terminated, nil
	}
	if err := d.finalizer.Remove(ctx, ps); err != nil {
		return ps.Status.Phase, fmt.Errorf("unable to removing finalizer, error %w", err)
	}
	d.recorder.Eventf(ps, v1.EventTypeNormal, "Terminated", "Prometheus Server Namespace %s Name %s Terminated", ps.Namespace, ps.Name)

	return v1alpha1.Terminated, nil
}

// Handlers return creation status handlers
func (d *deleter) Handlers() map[string]service.StateHandler {
	return map[string]service.StateHandler{
		v1alpha1.Terminating: d.Terminating,
	}
}
