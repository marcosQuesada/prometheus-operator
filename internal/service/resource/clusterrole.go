package resource

import (
	"context"
	"fmt"
	service2 "github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	rbac "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/rbac/v1"
)

const clusterRoleName = service2.MonitoringName + "-role"
const clusterRoleResourceName = "clusterroles"

type clusterRole struct {
	client kubernetes.Interface
	lister listersV1.ClusterRoleLister
	name   string
}

// NewClusterRole instantiates cluster role resource enforcer
func NewClusterRole(cl kubernetes.Interface, l listersV1.ClusterRoleLister) service2.ResourceEnforcer {
	return &clusterRole{
		client: cl,
		lister: l,
		name:   clusterRoleName,
	}
}

// EnsureCreation checks cluster role existence, if it's not found it will create it
func (c *clusterRole) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get cluster role %w", err)
	}

	return nil
}

// EnsureDeletion checks cluster role existence, if it's it will delete it
func (c *clusterRole) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("removing cluster role %s", c.name)
	err := c.client.RbacV1().ClusterRoles().Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to delete cluster role, error %w", err)
	}
	return nil
}

// IsCreated check if resource exists
func (c *clusterRole) IsCreated() (bool, error) {
	_, err := c.lister.Get(c.name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to get cluster role %w", err)
	}

	return true, nil
}

// Name returns resource enforcer target name
func (c *clusterRole) Name() string {
	return clusterRoleResourceName
}

func (c *clusterRole) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("creating cluster role %s", c.name)
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
		return fmt.Errorf("unable to create cluster role, error %w", err)
	}
	return nil
}
