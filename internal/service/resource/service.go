package resource

import (
	"context"
	"fmt"

	svc "github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/core/v1"
)

const prometheusServiceName = svc.MonitoringName + "-service"
const serviceResourceName = "services"

type service struct {
	client    kubernetes.Interface
	lister    listersV1.ServiceLister
	namespace string
	name      string
}

// NewService instantiates prometheus service resource enforcer
func NewService(cl kubernetes.Interface, l listersV1.ServiceLister) svc.ResourceEnforcer {
	return &service{
		client:    cl,
		lister:    l,
		namespace: svc.MonitoringNamespace,
		name:      prometheusServiceName,
	}
}

// EnsureCreation checks service existence, if it's not found it will create it
func (c *service) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Services(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get service %w", err)
	}

	return nil
}

// EnsureDeletion checks service existence, if it's it will delete it
func (c *service) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("removing service  %s", c.name)
	err := c.client.CoreV1().Services(c.namespace).Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to delete service, error %w", err)
	}
	return nil
}

// IsCreated check if resource exists
func (c *service) IsCreated() (bool, error) {
	_, err := c.lister.Services(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to get service %w", err)
	}

	return true, nil
}

// Name returns resource enforcer target name
func (c *service) Name() string {
	return serviceResourceName
}

func (c *service) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("creating service  %s", c.name)
	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      prometheusServiceName,
			Namespace: svc.MonitoringNamespace,
			Annotations: map[string]string{
				"prometheus.io/scrape": "true", // Prometheus service scrapped by itself
				"prometheus.io/port":   fmt.Sprintf("%d", prometheusHttpPort),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       prometheusServiceHttpPort,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(prometheusHttpPort),
				},
			},
			Selector: map[string]string{"app": svc.MonitoringName},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	_, err := c.client.CoreV1().Services(c.namespace).Create(ctx, s, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create service, error %w", err)
	}
	return nil
}
