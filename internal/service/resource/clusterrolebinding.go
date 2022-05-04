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

const clusterRoleBindingResourceName = "clusterrolebindings"
const clusterRoleBindingName = service2.MonitoringName + "-role-binding"
const rbacApiGroup = "rbac.authorization.k8s.io"
const serviceAccount = "ServiceAccount"

type clusterRoleBinding struct {
	client kubernetes.Interface
	lister listersV1.ClusterRoleBindingLister
	name   string
}

// NewClusterRoleBinding instantiates cluster role binding resource enforcer
func NewClusterRoleBinding(cl kubernetes.Interface, l listersV1.ClusterRoleBindingLister) service2.ResourceEnforcer {
	return &clusterRoleBinding{
		client: cl,
		lister: l,
		name:   clusterRoleBindingName,
	}
}

// EnsureCreation checks cluster role binding existence, if it's not found it will create it
func (c *clusterRoleBinding) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get cluster role bindings %w", err)
	}

	return nil
}

// EnsureDeletion checks cluster role binding existence, if it's it will delete it
func (c *clusterRoleBinding) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("removing cluster role binding %s", c.name)
	err := c.client.RbacV1().ClusterRoleBindings().Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to delete cluster role bindings, error %w", err)
	}
	return nil
}

// IsCreated check if resource exists
func (c *clusterRoleBinding) IsCreated() (bool, error) {
	_, err := c.lister.Get(c.name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to get cluster role binding %w", err)
	}

	return true, nil
}

// Name returns resource enforcer target name
func (c *clusterRoleBinding) Name() string {
	return clusterRoleBindingResourceName
}

func (c *clusterRoleBinding) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Debugf("creating cluster role binding %s", c.name)
	cm := &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.name,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbacApiGroup,
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      serviceAccount,
				Name:      "default",
				Namespace: service2.MonitoringNamespace,
			},
		},
	}

	_, err := c.client.RbacV1().ClusterRoleBindings().Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create cluster role binding, error %w", err)
	}
	return nil
}
