package service

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Finalizer takes care on add/remove finalizer to crd
type Finalizer interface {
	Ensure(ctx context.Context, ps *v1alpha1.PrometheusServer) error
	Add(ctx context.Context, ps *v1alpha1.PrometheusServer) error
	Remove(ctx context.Context, ps *v1alpha1.PrometheusServer) error
}

type finalizer struct {
	client versioned.Interface
}

// NewFinalizer instantiates finalizer
func NewFinalizer(c versioned.Interface) Finalizer {
	return &finalizer{client: c}
}

// Ensure check if crd has finalizer, if not it adds it
func (o *finalizer) Ensure(ctx context.Context, ps *v1alpha1.PrometheusServer) error {
	if HasFinalizer(ps) {
		return nil
	}

	if err := o.Add(ctx, ps); err != nil {
		return fmt.Errorf("unable to add finalizer, error %w", err)
	}
	return nil
}

// Add finalizer to crd
func (o *finalizer) Add(ctx context.Context, ps *v1alpha1.PrometheusServer) error {
	log.Debugf("Adding Finalizer from crd %s", ps.Name)
	p := ps.DeepCopy()
	p.Finalizers = append(p.Finalizers, v1alpha1.Name)
	_, err := o.client.K8slabV1alpha1().PrometheusServers(ps.Namespace).Update(ctx, p, metav1.UpdateOptions{})

	return err
}

// Remove finalizer from CRD
func (o *finalizer) Remove(ctx context.Context, ps *v1alpha1.PrometheusServer) error {
	log.Debugf("Removing Finalizer from crd %s", ps.Name)
	newFinalizers := []string{}
	for _, finalizer := range ps.Finalizers {
		if finalizer == v1alpha1.Name {
			continue
		}
		newFinalizers = append(newFinalizers, finalizer)
	}
	p := ps.DeepCopy()
	p.Finalizers = newFinalizers
	_, err := o.client.K8slabV1alpha1().PrometheusServers(ps.Namespace).Update(ctx, p, metav1.UpdateOptions{})

	return err
}

// HasFinalizer check if prometheusServer label is added as finalizer
func HasFinalizer(ps *v1alpha1.PrometheusServer) bool {
	if len(ps.Finalizers) == 0 {
		return false
	}

	for _, v := range ps.Finalizers {
		if v == v1alpha1.Name {
			return true
		}
	}
	return false
}
