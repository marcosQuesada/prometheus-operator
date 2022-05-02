package crd

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type builder struct {
	initializer Initializer
}

func NewBuilder(i Initializer) *builder {
	return &builder{
		initializer: i,
	}
}

func (b *builder) Create(ctx context.Context) error {
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
						//{
						//	Name:     "Config",
						//	Type:     "string",
						//	JSONPath: ".spec.config",
						//},
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

func (b *builder) IsAccepted(ctx context.Context) (bool, error) {
	return b.initializer.IsAccepted(ctx, v1alpha1.Name)
}

func (b *builder) EnsureCRDRegistered() error {
	acc, err := b.IsAccepted(context.Background())
	if err != nil {
		return fmt.Errorf("unable to check swarm crd status, error %v", err)
	}
	if acc {
		return nil
	}

	if err := b.Create(context.Background()); err != nil {
		return fmt.Errorf("unable to initialize swarm crd, error %v", err)
	}

	return nil
}
