package service

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// ResourceManager delegates responsibility on resource creation/removal
type ResourceManager interface {
	AllCreated() (bool, error)
	AllRemoved() (bool, error)
	CreateAll(ctx context.Context, p *v1alpha1.PrometheusServer) error
	DeleteAll(ctx context.Context, p *v1alpha1.PrometheusServer) error
}

// ResourceEnforcer taks care on resource creation/deletion
type ResourceEnforcer interface {
	EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error
	EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error
	IsCreated() (bool, error)
	Name() string
}

type resource struct {
	builders []ResourceEnforcer
}

// NewResource instantiates default resource manager
func NewResource(b ...ResourceEnforcer) ResourceManager {
	return &resource{
		builders: b,
	}
}

// AllCreated checks all resources exists
func (o *resource) AllCreated() (bool, error) {
	return o.allResourcesExist(true)
}

// AllRemoved checks none resources exists
func (o *resource) AllRemoved() (bool, error) {
	return o.allResourcesExist(false)
}

// CreateAll executes resource creation
func (o *resource) CreateAll(ctx context.Context, p *v1alpha1.PrometheusServer) error {
	log.Infof("Creating resources from prometheus server on namespace %s name %s ", p.Namespace, p.Name)

	for _, r := range o.builders {
		if err := r.EnsureCreation(ctx, p); err != nil {
			return fmt.Errorf("unable to ensure creation on %s error %w", r.Name(), err)
		}
	}

	return nil
}

// DeleteAll executes resource deletion
func (o *resource) DeleteAll(ctx context.Context, p *v1alpha1.PrometheusServer) error {
	log.Infof("Delete resources from prometheus server on namespace %s name %s ", p.Namespace, p.Name)

	// Running them from the end to beginning (Opposite as creation)
	for i := len(o.builders) - 1; i >= 0; i-- {
		r := o.builders[i]
		if err := r.EnsureDeletion(ctx, p); err != nil {
			return fmt.Errorf("unable to ensure creation on %s error %w", r.Name(), err)
		}
	}

	return nil
}

func (o *resource) allResourcesExist(mustExist bool) (bool, error) {
	for _, r := range o.builders {
		ok, err := r.IsCreated()
		if err != nil {
			return false, fmt.Errorf("resource %s creation check error %w", r.Name(), err)
		}
		if mustExist != ok {
			return false, nil
		}
	}

	return true, nil
}
