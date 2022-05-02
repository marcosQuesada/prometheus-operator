package crd

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"sync"
	"unicode"
)

type controller interface {
	Update(ctx context.Context, old, new *v1alpha1.PrometheusServer) error
	Delete(ctx context.Context, obj *v1alpha1.PrometheusServer) error
}

// Handler handles prometheus server state updates
type Handler struct {
	controller controller
	mutex      sync.RWMutex
}

// NewHandler instantiates prometheus server handler
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

	// @TODO: REMOVE IT
	diff := cmp.Diff(old, nsw)
	cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})
	fmt.Printf("UPDATE %s diff: %s \n", o.GetObjectKind(), cleanDiff)

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
