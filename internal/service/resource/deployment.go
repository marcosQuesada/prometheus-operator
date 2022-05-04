package resource

import (
	"context"
	"fmt"
	service2 "github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/apps/v1"
)

const prometheusDeploymentName = service2.MonitoringName + "-deployment"
const deploymentResourceName = "deployments"
const prometheusHttpPort = 9090
const prometheusServiceHttpPort = 8080
const prometheusConfigPath = "/etc/prometheus/"
const prometheusStoragePath = "/prometheus/"
const prometheusStorageVolumeName = "prometheus-storage-volume"
const prometheusConfigVolumeName = "prometheus-config-volume"
const prometheusReadinessEndpoint = "/-/ready"
const prometheusLivenessEndpoint = "/-/healthy"
const defaultInitialDelaySeconds = 2
const defaultTimeoutSeconds = 5

var prometheusConfigFileArg = fmt.Sprintf("--config.file=%sprometheus.yml", prometheusConfigPath)
var prometheusDbPathArg = fmt.Sprintf("--storage.tsdb.path=%s", prometheusStoragePath)

type deployment struct {
	client    kubernetes.Interface
	lister    listersV1.DeploymentLister
	namespace string
	name      string
}

// NewDeployment instantiates prometheus deployment resource enforcer
func NewDeployment(cl kubernetes.Interface, l listersV1.DeploymentLister) service2.ResourceEnforcer {
	return &deployment{
		client:    cl,
		lister:    l,
		namespace: service2.MonitoringNamespace,
		name:      prometheusDeploymentName,
	}
}

// EnsureCreation checks deployment existence, if it's not found it will create it
func (c *deployment) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Deployments(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get deployment %w", err)
	}

	return nil
}

// EnsureDeletion checks deployment existence, if it's it will delete it
func (c *deployment) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("removing deployment  %s", c.name)
	err := c.client.AppsV1().Deployments(c.namespace).Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to delete deployment, error %w", err)
	}
	return nil
}

// IsCreated check if resource exists
func (c *deployment) IsCreated() (bool, error) {
	_, err := c.lister.Deployments(c.namespace).Get(c.name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to get deployment %w", err)
	}

	return true, nil
}

// Name returns resource enforcer target name
func (c *deployment) Name() string {
	return deploymentResourceName
}

func (c *deployment) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("creating deployment  %s", c.name)
	replicas := int32(1)
	defaultPermission := int32(420)
	cm := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
			Labels:    map[string]string{"app": service2.MonitoringName},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": service2.MonitoringName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      c.name,
					Namespace: c.namespace,
					Labels:    map[string]string{"app": service2.MonitoringName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  service2.MonitoringName,
							Image: fmt.Sprintf("prom/prometheus:%s", obj.Spec.Version),
							Args: []string{
								prometheusConfigFileArg,
								prometheusDbPathArg,
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: prometheusHttpPort,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      prometheusConfigVolumeName,
									MountPath: prometheusConfigPath,
								},
								{
									Name:      prometheusStorageVolumeName,
									MountPath: prometheusStoragePath,
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
									Path: prometheusLivenessEndpoint,
									Port: intstr.FromInt(prometheusHttpPort),
								}},
								InitialDelaySeconds: defaultInitialDelaySeconds,
								TimeoutSeconds:      defaultTimeoutSeconds,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
									Path: prometheusReadinessEndpoint,
									Port: intstr.FromInt(prometheusHttpPort),
								}},
								InitialDelaySeconds: defaultInitialDelaySeconds,
								TimeoutSeconds:      defaultTimeoutSeconds,
							},
						}},
					Volumes: []corev1.Volume{
						{
							Name: prometheusConfigVolumeName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: prometheusConfigMapName},
									DefaultMode:          &defaultPermission,
								},
							},
						},
						{
							Name: prometheusStorageVolumeName,
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
		return fmt.Errorf("unable to create deployment, error %w", err)
	}
	return nil
}

func getImageName(version string) string {
	return fmt.Sprintf("prom/prometheus:%s", version)
}
