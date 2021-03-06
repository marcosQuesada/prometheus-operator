package usecase

import (
	"context"

	"github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

type reloader struct {
	generation service.Cache
	resource   service.ResourceManager
	recorder   record.EventRecorder
}

// NewReloader instantiates reloader use case status handlers
func NewReloader(c service.Cache, r service.ResourceManager, e record.EventRecorder) service.ConciliatorHandler {
	return &reloader{
		generation: c,
		resource:   r,
		recorder:   e,
	}
}

// Running Status handler
func (r *reloader) Running(ctx context.Context, ps *v1alpha1.PrometheusServer) (string, error) {
	defer runningProcessed.Inc()

	g := r.generation.Get(ps.Namespace, ps.Name)
	log.Infof("Prometheus Server on Running state with generation %d registered is on %d", ps.Generation, g)
	if g == ps.Generation || g == 0 {
		r.recorder.Eventf(ps, v1.EventTypeNormal, "Running", "Prometheus Server Namespace %s Name %s running", ps.Namespace, ps.Name)
		return ps.Status.Phase, nil
	}
	r.recorder.Eventf(ps, v1.EventTypeNormal, "Reloading", "Prometheus Server Namespace %s Name %s reloading", ps.Namespace, ps.Name)

	return v1alpha1.Reloading, nil
}

// Reloading Status handler
func (r *reloader) Reloading(ctx context.Context, ps *v1alpha1.PrometheusServer) (string, error) {
	defer reloadingProcessed.Inc()

	if err := r.resource.DeleteAll(ctx, ps); err != nil {
		r.recorder.Eventf(ps, v1.EventTypeWarning, "DeleteAllError", "error %v deleting resources", err.Error())

		return ps.Status.Phase, err
	}
	return v1alpha1.WaitingRemoval, nil
}

// WaitingRemoval Status handler
func (r *reloader) WaitingRemoval(ctx context.Context, ps *v1alpha1.PrometheusServer) (string, error) {
	defer waitingRemovalProcessed.Inc()

	ok, err := r.resource.AllRemoved()
	if err != nil {
		return ps.Status.Phase, err
	}
	if !ok {
		return ps.Status.Phase, nil
	}

	r.recorder.Eventf(ps, v1.EventTypeNormal, "Rebuilding", "Prometheus Server Namespace %s Name %s rebuilding", ps.Namespace, ps.Name)

	return v1alpha1.Initializing, nil
}

// Handlers return creation status handlers
func (r *reloader) Handlers() map[string]service.StateHandler {
	return map[string]service.StateHandler{
		v1alpha1.Running:        r.Running,
		v1alpha1.Reloading:      r.Reloading,
		v1alpha1.WaitingRemoval: r.WaitingRemoval,
	}
}
