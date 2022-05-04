package resource

import (
	"context"
	"fmt"
	service2 "github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/core/v1"
)

const prometheusConfigMapName = service2.MonitoringName + "-config"
const prometheusConfigMapKey = "prometheus.yml"
const configMapResourceName = "configmaps"

type configMap struct {
	client    kubernetes.Interface
	lister    listersV1.ConfigMapLister
	namespace string
	name      string
}

// NewConfigMap instantiates configmap resource enforcer
func NewConfigMap(cl kubernetes.Interface, l listersV1.ConfigMapLister) service2.ResourceEnforcer {
	return &configMap{
		client:    cl,
		lister:    l,
		namespace: service2.MonitoringNamespace,
		name:      prometheusConfigMapName,
	}
}

// EnsureCreation checks configmap existence, if it's not found it will create it
func (c *configMap) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.ConfigMaps(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get config map %w", err)
	}

	return nil
}

// EnsureDeletion checks configmap existence, if it's it will delete it
func (c *configMap) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("removing configmap  %s", c.name)
	err := c.client.CoreV1().ConfigMaps(c.namespace).Delete(ctx, prometheusConfigMapName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to delete configmap, error %w", err)
	}
	return nil
}

// IsCreated check if resource exists
func (c *configMap) IsCreated() (bool, error) {
	_, err := c.lister.ConfigMaps(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to get cofigmap %w", err)
	}

	return true, nil
}

// Name returns resource enforcer target name
func (c *configMap) Name() string {
	return configMapResourceName
}

func (c *configMap) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("creating configmap  %s", c.name)
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
		},
		Data: map[string]string{prometheusConfigMapKey: obj.Spec.Config},
	}
	_, err := c.client.CoreV1().ConfigMaps(c.namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create configmap, error %w", err)
	}
	return nil
}
