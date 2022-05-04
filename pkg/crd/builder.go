package crd

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

const pollFrequency = time.Second
const pollTimeout = 10 * time.Second

type manager struct {
	apiExtensionsClientSet apiextensionsclientset.Interface
}

// NewManager instantiates initializer
func NewManager(api apiextensionsclientset.Interface) Initializer {
	return &manager{apiExtensionsClientSet: api}
}

// Create will check CRD existence, if not found it will create it and wait until accepted
func (c *manager) Create(ctx context.Context, cr *v1.CustomResourceDefinition) error {
	_, err := c.apiExtensionsClientSet.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, cr, metav1.CreateOptions{})

	if err != nil {
		log.Fatalf("unable to create CRD , error: %v", err.Error())
	}

	return c.waitCRDAccepted(ctx, cr.Name)
}

// IsAccepted checks if CRD is accepted
func (c *manager) IsAccepted(ctx context.Context, resourceName string) (bool, error) {
	cr, err := c.apiExtensionsClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, resourceName, metav1.GetOptions{})
	if _, ok := err.(*apiErrors.StatusError); ok {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	for _, condition := range cr.Status.Conditions {
		if condition.Type == v1.Established &&
			condition.Status == v1.ConditionTrue {
			return true, nil
		}
	}

	return false, fmt.Errorf("CRD %s is not accepted", resourceName)
}

func (c *manager) waitCRDAccepted(ctx context.Context, resourceName string) error {
	err := wait.Poll(pollFrequency, pollTimeout, func() (bool, error) {
		return c.IsAccepted(ctx, resourceName)
	})

	return err
}
