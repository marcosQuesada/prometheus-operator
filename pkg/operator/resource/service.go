package resource

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/core/v1"
)

const prometheusServiceName = operator.MonitoringName + "-service"

type service struct {
	client    kubernetes.Interface
	lister    listersV1.ServiceLister
	namespace string
	name      string
}

func NewService(cl kubernetes.Interface, l listersV1.ServiceLister) *service {
	return &service{
		client:    cl,
		lister:    l,
		namespace: operator.MonitoringNamespace,
		name:      prometheusServiceName,
	}
}

func (c *service) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Services(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}

	return nil
}

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

func (c *service) Name() string {
	return "service"
}

func (c *service) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("creating service  %s", c.name)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      prometheusServiceName,
			Namespace: operator.MonitoringNamespace,
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "9090",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(9090),
					NodePort:   30000,
				},
			},
			Selector: map[string]string{"app": "prometheus-server"}, // @TODO
			Type:     corev1.ServiceTypeNodePort,
		},
	}
	_, err := c.client.CoreV1().Services(c.namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create service, error %w", err)
	}
	return nil
}
