package resource

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	log "github.com/sirupsen/logrus"
	rbac "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listersV1 "k8s.io/client-go/listers/rbac/v1"
)

const clusterRoleBindingName = operator.MonitoringName + "-role-binding"

type clusterRoleBinding struct {
	client kubernetes.Interface
	lister listersV1.ClusterRoleBindingLister
	name   string
}

func NewClusterRoleBinding(cl kubernetes.Interface, l listersV1.ClusterRoleBindingLister) *clusterRoleBinding {
	return &clusterRoleBinding{
		client: cl,
		lister: l,
		name:   clusterRoleBindingName,
	}
}

func (c *clusterRoleBinding) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	_, err := c.lister.Get(c.name)
	if apierrors.IsNotFound(err) {
		return c.create(ctx, obj)
	}

	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}

	return nil
}

func (c *clusterRoleBinding) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("removing cluster role binding %s", c.name)
	err := c.client.RbacV1().ClusterRoleBindings().Delete(ctx, c.name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to delete configmap, error %w", err)
	}
	return nil
}

func (c *clusterRoleBinding) Name() string {
	return "clusterrolebinding"
}

func (c *clusterRoleBinding) create(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	log.Infof("creating cluster role binding %s", c.name)
	cm := &rbac.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: c.name,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName, // @TODO: SURE?
		},
		Subjects: []rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: operator.MonitoringNamespace,
			},
		},
	}

	_, err := c.client.RbacV1().ClusterRoleBindings().Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create configmap, error %w", err)
	}
	return nil
}
