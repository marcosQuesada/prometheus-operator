package crd

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Initializer iw responsible on ensuring CRD is registered on api-server
type Initializer interface {
	Create(ctx context.Context, cr *v1.CustomResourceDefinition) error
	IsAccepted(ctx context.Context, resourceName string) (bool, error)
}

type Builder struct {
	initializer Initializer
}

// NewBuilder its responsible for initialize CRD on cluster
func NewBuilder(i Initializer) *Builder {
	return &Builder{
		initializer: i,
	}
}

// EnsureCRDRegistration ensures CRD is created, if it didn't it will force creation
func (b *Builder) EnsureCRDRegistration(ctx context.Context) error {
	log.Info("Ensuring crd is registered")
	acc, err := b.initializer.IsAccepted(ctx, v1alpha1.Name)
	if err != nil {
		return fmt.Errorf("unable to check crd status, error %w", err)
	}

	if acc {
		return nil
	}

	if err := b.create(context.Background()); err != nil {
		return fmt.Errorf("unable to initialize crd, error %w", err)
	}

	return nil
}

// Create defines PrometheusServer CRD resource
func (b *Builder) create(ctx context.Context) error {
	log.Info("Creating Prometheus Server CRD")
	cr := &v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.Name,
		},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: v1alpha1.GroupName,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name:    v1alpha1.Version,
					Served:  true,
					Storage: true,
					Subresources: &v1.CustomResourceSubresources{
						Status: &v1.CustomResourceSubresourceStatus{},
					},
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]v1.JSONSchemaProps{
										"version": {Type: "string"},
										"config":  {Type: "string"},
									},
									Required: []string{"version", "config"},
								},
								"status": {
									Type: "object",
									Properties: map[string]v1.JSONSchemaProps{
										"phase": {
											Type: "string",
										},
									},
								},
							},
						},
					},
					AdditionalPrinterColumns: []v1.CustomResourceColumnDefinition{
						{
							Name:     "Version",
							Type:     "string",
							JSONPath: ".spec.version",
						},
						{
							Name:     "Age",
							Type:     "date",
							JSONPath: ".metadata.creationTimestamp",
						},
						{
							Name:     "Status",
							Type:     "string",
							JSONPath: ".status.phase",
						},
					},
				},
			},
			Scope: v1.NamespaceScoped,
			Names: v1.CustomResourceDefinitionNames{
				Plural:     v1alpha1.Plural,
				Singular:   v1alpha1.Singular,
				Kind:       v1alpha1.CrdKind,
				ShortNames: []string{v1alpha1.ShortName},
			},
		},
	}

	return b.initializer.Create(ctx, cr)
}
