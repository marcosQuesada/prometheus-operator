package resource

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	service2 "github.com/marcosQuesada/prometheus-operator/pkg/service"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/core/v1"
)

type namespace struct {
	client kubernetes.Interface
	lister listersV1.NamespaceLister
	name   string
}

func NewNamespace(cl kubernetes.Interface, l listersV1.NamespaceLister) *namespace {
	return &namespace{
		client: cl,
		lister: l,
		name:   service2.MonitoringNamespace,
	}
}

func (c *namespace) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}

	return nil
}

func (c *namespace) Name() string {
	return "namespace"
}

func (c *namespace) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("removing namespace  %s", c.name)
	err := c.client.CoreV1().Namespaces().Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to delete configmap, error %w", err)
	}
	return nil
}

func (c *namespace) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("creating namespace  %s", c.name)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: service2.MonitoringNamespace,
		},
	}
	_, err := c.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create service, error %w", err)
	}
	return nil
}
