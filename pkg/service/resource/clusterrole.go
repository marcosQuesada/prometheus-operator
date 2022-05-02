package resource

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	service2 "github.com/marcosQuesada/prometheus-operator/pkg/service"
	log "github.com/sirupsen/logrus"
	rbac "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/rbac/v1"
)

const clusterRoleName = service2.MonitoringName + "-role"

type clusterRole struct {
	client kubernetes.Interface
	lister listersV1.ClusterRoleLister
	name   string
}

func NewClusterRole(cl kubernetes.Interface, l listersV1.ClusterRoleLister) *clusterRole {
	return &clusterRole{
		client: cl,
		lister: l,
		name:   clusterRoleName,
	}
}

func (c *clusterRole) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}

	return nil
}

func (c *clusterRole) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("removing cluster role %s", c.name)
	err := c.client.RbacV1().ClusterRoles().Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to delete configmap, error %w", err)
	}
	return nil
}

func (c *clusterRole) Name() string {
	return "clusterrole"
}

func (c *clusterRole) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("creating cluster role %s", c.name)
	cm := &rbac.ClusterRole{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: c.name,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{
					"nodes",
					"nodes/proxy",
					"services",
					"endpoints",
					"pods",
				},
				Verbs: []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{
					"ingress",
				},
				Verbs: []string{"get", "list", "watch"},
			},
			{
				NonResourceURLs: []string{"/metrics"},
				Verbs:           []string{"get"},
			},
		},
	}

	_, err := c.client.RbacV1().ClusterRoles().Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create configmap, error %w", err)
	}
	return nil
}
