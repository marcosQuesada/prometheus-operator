package resource

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/apps/v1"
)

const prometheusDeploymentName = operator.MonitoringName

type deployment struct {
	client    kubernetes.Interface
	lister    listersV1.DeploymentLister
	namespace string
	name      string
}

func NewDeployment(cl kubernetes.Interface, l listersV1.DeploymentLister) *deployment {
	return &deployment{
		client:    cl,
		lister:    l,
		namespace: operator.MonitoringNamespace,
		name:      prometheusDeploymentName + "-deployment",
	}
}

func (c *deployment) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Deployments(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}

	return nil
}

func (c *deployment) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("removing deployment  %s", c.name)
	err := c.client.AppsV1().Deployments(c.namespace).Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to delete configmap, error %w", err)
	}
	return nil
}

func (c *deployment) Name() string {
	return "deployment"
}

func (c *deployment) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("creating deployment  %s", c.name)
	replicas := int32(1)
	defaultPermission := int32(420)
	cm := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
			Labels:    map[string]string{"app": operator.MonitoringName},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": operator.MonitoringName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      c.name,
					Namespace: c.namespace,
					Labels:    map[string]string{"app": operator.MonitoringName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  operator.MonitoringName,
							Image: "prom/prometheus", // @TODO: Add version
							Args: []string{
								"--config.file=/etc/prometheus/prometheus.yml",
								"--storage.tsdb.path=/prometheus/",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 9090,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "prometheus-config-volume",
									MountPath: "/etc/prometheus/",
								},
								{
									Name:      "prometheus-storage-volume",
									MountPath: "/prometheus/",
								},
							},
						}},
					Volumes: []corev1.Volume{
						{
							Name: "prometheus-config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: prometheusConfigMapName},
									DefaultMode:          &defaultPermission,
								},
							},
						},
						{
							Name: "prometheus-storage-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{},
	}
	_, err := c.client.AppsV1().Deployments(c.namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create configmap, error %w", err)
	}
	return nil
}
