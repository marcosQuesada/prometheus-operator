package resource

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	svc "github.com/marcosQuesada/prometheus-operator/pkg/service"
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
	s, err := c.lister.Services(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}

	log.Infof("Ensurecreation without errors, conditions %v", s.Status.Conditions) // @TODO
	return nil
}

// EnsureDeletion checks service existence, if it's it will delete it
func (c *service) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("removing service  %s", c.name)
	err := c.client.CoreV1().Services(c.namespace).Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to delete configmap, error %w", err)
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
		return false, fmt.Errorf("unable to get service %v", err)
	}

	return true, nil
}

// Name returns resource enforcer target name
func (c *service) Name() string {
	return serviceResourceName
}

func (c *service) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("creating service  %s", c.name)
	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      prometheusServiceName,
			Namespace: svc.MonitoringNamespace,
			Annotations: map[string]string{
				"prometheus.io/scrape": "true", // Prometheus service scrapped by itself // @TODO
				"prometheus.io/port":   "9090",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       8080, // @TODO: TO config!!!
					TargetPort: intstr.FromInt(9090),
					NodePort:   30000, // @TODO: Reconsider
				},
			},
			Selector: map[string]string{"app": svc.MonitoringName},
			Type:     corev1.ServiceTypeNodePort,
		},
	}
	_, err := c.client.CoreV1().Services(c.namespace).Create(ctx, s, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create service, error %w", err)
	}
	return nil
}
