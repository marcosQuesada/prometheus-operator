package usecase

import (
	"context"

	"github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

type creator struct {
	finalizer service.Finalizer
	resource  service.ResourceManager
	recorder  record.EventRecorder
}

// NewCreator instantiates creation use case states
func NewCreator(f service.Finalizer, r service.ResourceManager, e record.EventRecorder) service.ConciliatorHandler {
	return &creator{
		finalizer: f,
		resource:  r,
		recorder:  e,
	}
}

// Empty Status handler
func (c *creator) Empty(ctx context.Context, ps *v1alpha1.PrometheusServer) (newStatus string, err error) {
	defer emptyProcessed.Inc()
	if !service.HasFinalizer(ps) {
		if err := c.finalizer.Add(ctx, ps); err != nil {
			return ps.Status.Phase, err
		}
		return ps.Status.Phase, nil
	}

	return v1alpha1.Initializing, nil
}

// Initializing Status handler
func (c *creator) Initializing(ctx context.Context, ps *v1alpha1.PrometheusServer) (newStatus string, err error) {
	defer initializingProcessed.Inc()

	if err := c.resource.CreateAll(ctx, ps); err != nil {
		c.recorder.Eventf(ps, v1.EventTypeWarning, "createAllError", "error %v creating resources", err.Error())
		return ps.Status.Phase, err
	}
	c.recorder.Event(ps, v1.EventTypeNormal, "createAllSuccess", "resources created with success")
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
func (c *creator) Handlers() map[string]service.StateHandler {
	return map[string]service.StateHandler{
		v1alpha1.Empty:           c.Empty,
		v1alpha1.Initializing:    c.Initializing,
		v1alpha1.WaitingCreation: c.WaitingCreation,
	}
}
