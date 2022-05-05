package resource

import (
	"context"
	"testing"
	"time"

	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	k8stest "k8s.io/client-go/testing"
)

func TestItCreatesDeploymentOnCreationRequest(t *testing.T) {
	clientSet := fake.NewSimpleClientset()
	sif := informers.NewSharedInformerFactory(clientSet, 0)
	i := sif.Apps().V1().Deployments()

	svc := NewDeployment(clientSet, i.Lister())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	version := "v1.0.1"
	pm := &v1alpha1.PrometheusServer{
		Spec: v1alpha1.PrometheusServerSpec{Version: version},
	}
	if err := svc.EnsureCreation(ctx, pm); err != nil {
		t.Fatalf("unable to ensure deployment creation, error %v", err)
	}

	clActions := clientSet.Actions()
	if expected, got := 1, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	action := clActions[0]
	if expected, got := "create", action.GetVerb(); expected != got {
		t.Fatalf("unexpected verb, expected %s got %s", expected, got)
	}
	if expected, got := deploymentResourceName, action.GetResource().Resource; expected != got {
		t.Fatalf("unexpected resource, expected %s got %s", expected, got)
	}

	v, ok := action.(k8stest.CreateAction)
	if !ok {
		t.Fatalf("unexpected type got %T", action)
	}

	d, ok := v.GetObject().(*v1.Deployment)
	if !ok {
		t.Fatalf("unexpected type got %T", v.GetObject())
	}

	if expected, got := d.Spec.Template.Spec.Containers[0].Image, getImageName(version); expected != got {
		t.Fatalf("image version do not match, expected %s got %s", expected, got)
	}
}

func TestItRemovesDeploymentOnRemovalRequest(t *testing.T) {
	clientSet := fake.NewSimpleClientset()
	sif := informers.NewSharedInformerFactory(clientSet, 0)
	i := sif.Apps().V1().Deployments()

	svc := NewDeployment(clientSet, i.Lister())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	pm := &v1alpha1.PrometheusServer{}
	if err := svc.EnsureDeletion(ctx, pm); err != nil {
		t.Fatalf("unable to ensure deployment deletion, error %v", err)
	}
	clActions := clientSet.Actions()
	if expected, got := 1, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	action := clActions[0]
	if expected, got := "delete", action.GetVerb(); expected != got {
		t.Fatalf("unexpected verb, expected %s got %s", expected, got)
	}
}
