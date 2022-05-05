package crd

import (
	"context"
	"testing"

	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsFake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestItRecognizedCreatedCrdDevelopment(t *testing.T) {
	api := apiextensionsFake.NewSimpleClientset(&v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.Name,
		},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: v1alpha1.GroupName,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name: v1alpha1.Version,
				},
			},
		},
		Status: v1.CustomResourceDefinitionStatus{
			Conditions: []v1.CustomResourceDefinitionCondition{
				{
					Type:   v1.Established,
					Status: v1.ConditionTrue,
				},
			},
		},
	})

	m := NewManager(api)

	a, err := m.IsAccepted(context.Background(), v1alpha1.Name)
	if err != nil {
		t.Fatalf("unable check Is accepted on api server, error %v", err)
	}
	if !a {
		t.Error("expected accepted")
	}
}
