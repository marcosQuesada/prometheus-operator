package service

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned"
	v1alpha1Lister "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/listers/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MonitoringNamespace = "monitoring"
const MonitoringName = "prometheus-server"

type Conciliator interface {
	Conciliate(ctx context.Context, ps *v1alpha1.PrometheusServer) (string, error)
}

type Cache interface {
	Get(namespace, name string) int64
	Set(namespace, name string, v int64)
	Remove(namespace, name string)
}

type operator struct {
	lister          v1alpha1Lister.PrometheusServerLister
	client          versioned.Interface
	generationCache Cache
	conciliator     Conciliator
}

func NewOperator(l v1alpha1Lister.PrometheusServerLister, cl versioned.Interface, g Cache, c Conciliator) *operator {
	return &operator{
		lister:          l,
		client:          cl,
		generationCache: g,
		conciliator:     c,
	}
}

func (o *operator) Update(ctx context.Context, namespace, name string) error {
	defer updatesProcessed.Inc()

	ps, err := o.lister.PrometheusServers(namespace).Get(name)
	if err != nil {
		return fmt.Errorf("unable to get prometheus server  on namespace %s name %s definition , error %v", namespace, name, err)
	}
	log.Infof("Update called namespace %s name %s Status %s ", ps.Namespace, ps.Name, ps.Status.Phase)

	defer o.generationCache.Set(namespace, name, ps.Generation)

	if !ps.DeletionTimestamp.IsZero() && ps.Status.Phase != v1alpha1.Terminating {
		if err := o.updateStatus(ctx, ps, v1alpha1.Terminating); err != nil {
			return fmt.Errorf("unable to update status, error %v", err)
		}
		return nil
	}

	newState, err := o.conciliator.Conciliate(ctx, ps)
	if err != nil {
		return fmt.Errorf("unable to conciliate, error %v", err)
	}

	if ps.Status.Phase == newState || newState == v1alpha1.Terminated {
		return nil
	}

	log.Infof("Updating Status from %s to %s", ps.Status.Phase, newState)

	if err := o.updateStatus(ctx, ps, newState); err != nil {
		return fmt.Errorf("unable to update status, error %v", err)
	}
	return nil
}

func (o *operator) Delete(ctx context.Context, namespace, name string) error {
	defer deletesProcessed.Inc()

	log.Infof("Delete called namespace %s name %s ", namespace, name)
	o.generationCache.Remove(namespace, name)

	return nil
}

func (o *operator) updateStatus(ctx context.Context, ps *v1alpha1.PrometheusServer, status string) error {
	defer statusUpdatesProcessed.Inc()

	log.Infof("Updating status from crd %s to %s", ps.Name, status)
	ps.Status.Phase = status
	_, err := o.client.K8slabV1alpha1().PrometheusServers(ps.Namespace).UpdateStatus(ctx, ps, metav1.UpdateOptions{})
	return err
}
