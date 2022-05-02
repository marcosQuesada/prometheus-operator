package resource

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/core/v1"
)

const prometheusConfigMapName = operator.MonitoringName + "-config"
const prometheusConfigMapKey = "prometheus.yml"

type configMap struct {
	client    kubernetes.Interface
	lister    listersV1.ConfigMapLister
	namespace string
	name      string
}

func NewConfigMap(cl kubernetes.Interface, l listersV1.ConfigMapLister) *configMap {
	return &configMap{
		client:    cl,
		lister:    l,
		namespace: operator.MonitoringNamespace,
		name:      prometheusConfigMapName,
	}
}

func (c *configMap) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.ConfigMaps(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}

	return nil
}

func (c *configMap) Name() string {
	return "configmap" // @TODO: THIS as constant
}

func (c *configMap) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("removing configmap  %s", c.name)
	err := c.client.CoreV1().ConfigMaps(c.namespace).Delete(ctx, prometheusConfigMapName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to delete configmap, error %w", err)
	}
	return nil
}

func (c *configMap) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("creating configmap  %s", c.name)
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
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
