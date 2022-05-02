package crd

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
)

type controller interface {
	Update(ctx context.Context, old, new *v1alpha1.PrometheusServer) error
	Delete(ctx context.Context, obj *v1alpha1.PrometheusServer) error
}

// Handler handles swarm state updates
type Handler struct {
	controller controller
	mutex      sync.RWMutex
}

// NewHandler instantiates swarm handler
func NewHandler(c controller) *Handler {
	return &Handler{
		controller: c,
	}
}

func (h *Handler) Create(ctx context.Context, o runtime.Object) error {
	sw := o.(*v1alpha1.PrometheusServer)
	log.Infof("Create PrometheusServer Namespace %s name %s Version %s status %s", sw.Namespace, sw.Name, sw.Spec.Version, sw.Status)

	if err := h.controller.Update(ctx, nil, sw); err != nil {
		return fmt.Errorf("unable to process prometheus server creation %s %s error %v", sw.Namespace, sw.Name, err)
	}

	return nil
}

func (h *Handler) Update(ctx context.Context, o, n runtime.Object) error {
	old := o.(*v1alpha1.PrometheusServer)
	nsw := n.(*v1alpha1.PrometheusServer)
	log.Infof("Update PrometheusServer Namespace %s name %s Version%s status %s", old.Namespace, old.Name, old.Spec.Version, old.Status)

	// @TODO: Move to predicates, to allow update status propagation
	//if Equals(old, nsw) { // @TODO
	//	return nil
	//}

	if err := h.controller.Update(ctx, old, nsw); err != nil {
		return fmt.Errorf("unable to update  prometheus server %s %s error %v", nsw.Namespace, nsw.Name, err)
	}

	return nil
}

func (h *Handler) Delete(ctx context.Context, o runtime.Object) error {
	sw := o.(*v1alpha1.PrometheusServer)
	log.Infof("Delete PrometheusServer Namespace %s name %s", sw.Namespace, sw.Name)

	if err := h.controller.Delete(ctx, sw); err != nil {
		return fmt.Errorf("unable to delete  prometheus server %s %s error %v", sw.Namespace, sw.Name, err)
	}

	return nil
}

func Equals(o, n *v1alpha1.PrometheusServer) bool {
	if o.Spec.Config != n.Spec.Config {
		return false
	}

	if o.Spec.Version != n.Spec.Version {
		return false
	}

	return true
}
