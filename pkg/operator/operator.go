package operator

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MonitoringNamespace = "monitoring"
const MonitoringName = "prometheus-server"

type ResourceEnforcer interface {
	EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error
	EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error
	Name() string
}

type operator struct {
	client   versioned.Interface
	builders []ResourceEnforcer
}

func NewOperator(cl versioned.Interface, b []ResourceEnforcer) *operator {
	return &operator{
		client:   cl,
		builders: b,
	}
}

func (o *operator) Update(ctx context.Context, old, new *v1alpha1.PrometheusServer) error {
	log.Infof("Update called namespace %s name %s Status %s ", new.Namespace, new.Name, new.Status.Phase)
	if !new.DeletionTimestamp.IsZero() {
		return o.Delete(ctx, new)
	}

	// @TODO: Finalizers Exists ?
	if !new.HasFinalizer(v1alpha1.Name) {
		if err := o.addFinalizer(ctx, new); err != nil {
			return fmt.Errorf("unable to add finalizer, error %v", err)
		}
		return nil // @TODO: Let them to another update event
	}

	switch new.Status.Phase {
	case v1alpha1.Empty:
		if err := o.updateStatus(ctx, new, v1alpha1.Initializing); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	case v1alpha1.Initializing:
		log.Infof("Processing Initializing state %s ", new.Status.Phase)
		for _, r := range o.builders {
			if err := r.EnsureCreation(ctx, new); err != nil {
				return fmt.Errorf("unable to ensure creation on %s error %v", r.Name(), err)
			}
		}
		if err := o.updateStatus(ctx, new, v1alpha1.Waiting); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	case v1alpha1.Waiting:
		log.Infof("Processing state %s ", new.Status.Phase)
		for _, r := range o.builders {
			if err := r.EnsureCreation(ctx, new); err != nil {
				return fmt.Errorf("unable to ensure creation on %s error %v", r.Name(), err)
			}
		}
		if err := o.updateStatus(ctx, new, v1alpha1.Running); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	case v1alpha1.Running:
		// @TODO: Do nothing
	case v1alpha1.Terminating:
		// @TODO: Shouldn't happen here!

	}

	return nil
}

func (o *operator) Delete(ctx context.Context, p *v1alpha1.PrometheusServer) error {
	log.Infof("Delete called namespace %s name %s ", p.Namespace, p.Name)

	if p.Status.Phase != v1alpha1.Terminating {
		if err := o.updateStatus(ctx, p, v1alpha1.Terminating); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	}

	// @TODO: Running them from the end to beginning (Opposite as creation)
	for i := len(o.builders) - 1; i >= 0; i-- {
		r := o.builders[i]
		if err := r.EnsureDeletion(ctx, p); err != nil {
			return fmt.Errorf("unable to ensure creation on %s error %v", r.Name(), err)
		}
	}

	if p.HasFinalizer(v1alpha1.Name) {
		if err := o.removeFinalizer(ctx, p); err != nil {
			return fmt.Errorf("unable to removing finalizer, error %v", err)
		}
		return nil
	}
	return nil
}

// @TODO: Think on segregate it!
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

// @TODO: Move to predicates, to allow update status propagation
func Equals(o, n *v1alpha1.PrometheusServer) bool {
	if o.Spec.Config != n.Spec.Config {
		return false
	}

	if o.Spec.Version != n.Spec.Version {
		return false
	}

	return true
}
