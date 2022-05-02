package service

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned"
	v1alpha1Lister "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/listers/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

const MonitoringNamespace = "monitoring"
const MonitoringName = "prometheus-server"

type ResourceEnforcer interface {
	EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error
	EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error
	Name() string
}

type operator struct {
	lister     v1alpha1Lister.PrometheusServerLister
	client     versioned.Interface
	builders   []ResourceEnforcer
	generation map[string]int64 // track CRD generation to allow rollout on new versions
	mutex      sync.RWMutex
}

func NewOperator(l v1alpha1Lister.PrometheusServerLister, cl versioned.Interface, b []ResourceEnforcer) *operator {
	return &operator{
		lister:     l,
		client:     cl,
		builders:   b,
		generation: map[string]int64{},
	}
}

func (o *operator) Update(ctx context.Context, namespace, name string) error {
	ps, err := o.lister.PrometheusServers(namespace).Get(name)
	if err != nil {
		return fmt.Errorf("unable to get prometheus server  on namespace %s name %s definition , error %v", namespace, name, err)
	}
	log.Infof("Update called namespace %s name %s Status %s ", ps.Namespace, ps.Name, ps.Status.Phase)

	defer o.setGeneration(namespace, name, ps.Generation)

	if !ps.DeletionTimestamp.IsZero() { //  && ps.Status.Phase != v1alpha1.Terminating
		return o.Delete(ctx, namespace, name)
	}

	if !ps.HasFinalizer(v1alpha1.Name) {
		if err := o.addFinalizer(ctx, ps); err != nil {
			return fmt.Errorf("unable to add finalizer, error %v", err)
		}
		return nil
	}

	switch ps.Status.Phase {
	case v1alpha1.Empty:
		if err := o.updateStatus(ctx, ps, v1alpha1.Initializing); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	case v1alpha1.Initializing:
		log.Infof("Processing Initializing state %s ", ps.Status.Phase)
		for _, r := range o.builders {
			if err := r.EnsureCreation(ctx, ps); err != nil {
				return fmt.Errorf("unable to ensure creation on %s error %v", r.Name(), err)
			}
		}
		if err := o.updateStatus(ctx, ps, v1alpha1.Waiting); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	case v1alpha1.Waiting:
		log.Infof("Processing state %s", ps.Status.Phase)
		for _, r := range o.builders {
			if err := r.EnsureCreation(ctx, ps); err != nil {
				return fmt.Errorf("unable to ensure creation on %s error %v", r.Name(), err)
			}
		}
		if err := o.updateStatus(ctx, ps, v1alpha1.Running); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	case v1alpha1.Running:
		g := o.getGeneration(namespace, name)
		log.Infof("Update om Running state with generation %d registered is on %d", ps.Generation, g)
		if g == ps.Generation || g == 0 {
			return nil
		}

		if err := o.updateStatus(ctx, ps, v1alpha1.Reloading); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil

	case v1alpha1.Reloading:
		log.Infof("Reloading Stack %s", ps.Name)
		if err := o.delete(ctx, ps); err != nil {
			return err
		}

		if err := o.updateStatus(ctx, ps, v1alpha1.Initializing); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
	case v1alpha1.Terminating:
		// do nothing
	}

	return nil
}

func (o *operator) Delete(ctx context.Context, namespace, name string) error {
	log.Infof("Delete called namespace %s name %s ", namespace, name)
	defer o.removeGeneration(namespace, name)

	p, err := o.lister.PrometheusServers(namespace).Get(name)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to get prometheus server on namespace %s name %s definition , error %v", namespace, name, err)
	}

	if p.Status.Phase != v1alpha1.Terminating {
		if err := o.updateStatus(ctx, p, v1alpha1.Terminating); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	}

	if err := o.delete(ctx, p); err != nil {
		return err
	}

	if !p.HasFinalizer(v1alpha1.Name) {
		return nil
	}

	if err := o.removeFinalizer(ctx, p); err != nil {
		return fmt.Errorf("unable to removing finalizer, error %v", err)
	}

	return nil
}

func (o *operator) delete(ctx context.Context, p *v1alpha1.PrometheusServer) error {
	log.Infof("Delete called namespace %s name %s ", p.Namespace, p.Name)

	// Running them from the end to beginning (Opposite as creation)
	for i := len(o.builders) - 1; i >= 0; i-- {
		r := o.builders[i]
		if err := r.EnsureDeletion(ctx, p); err != nil {
			return fmt.Errorf("unable to ensure creation on %s error %v", r.Name(), err)
		}
	}

	return nil
}

func (o *operator) addFinalizer(ctx context.Context, ps *v1alpha1.PrometheusServer) error {
	log.Infof("Adding Finalizer from crd %s", ps.Name)
	ps.Finalizers = append(ps.Finalizers, v1alpha1.Name)
	_, err := o.client.K8slabV1alpha1().PrometheusServers(ps.Namespace).Update(ctx, ps, metav1.UpdateOptions{})

	return err
}

func (o *operator) removeFinalizer(ctx context.Context, ps *v1alpha1.PrometheusServer) error {
	log.Infof("Removing Finalizer from crd %s", ps.Name)
	newFinalizers := []string{}
	for _, finalizer := range ps.Finalizers {
		if finalizer == v1alpha1.Name {
			continue
		}
		newFinalizers = append(newFinalizers, finalizer)
	}
	ps.Finalizers = newFinalizers
	_, err := o.client.K8slabV1alpha1().PrometheusServers(ps.Namespace).Update(ctx, ps, metav1.UpdateOptions{})

	return err
}

func (o *operator) updateStatus(ctx context.Context, ps *v1alpha1.PrometheusServer, status string) error {
	log.Infof("Updating status from crd %s to %s", ps.Name, status)
	ps.Status.Phase = status
	_, err := o.client.K8slabV1alpha1().PrometheusServers(ps.Namespace).UpdateStatus(ctx, ps, metav1.UpdateOptions{})
	return err
}

func (o *operator) setGeneration(namespace, name string, v int64) {
	log.Infof("Setting to Registry %s/%s on generation %d", namespace, name, v)
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.generation[fmt.Sprintf("%s/%s", namespace, name)] = v
}

func (o *operator) removeGeneration(namespace, name string) {
	log.Infof("Removing from registry %s/%s", namespace, name)
	o.mutex.Lock()
	defer o.mutex.Unlock()

	delete(o.generation, fmt.Sprintf("%s/%s", namespace, name))
}

func (o *operator) getGeneration(namespace, name string) int64 {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	v, ok := o.generation[fmt.Sprintf("%s/%s", namespace, name)]
	if !ok {
		return 0
	}
	return v
}
