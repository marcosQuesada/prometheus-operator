package resource

import (
	"context"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestItCreatesServiceOnCreationRequest(t *testing.T) {
	clientSet := fake.NewSimpleClientset()

	sif := informers.NewSharedInformerFactory(clientSet, 0)
	i := sif.Core().V1().Services()

	svc := NewService(clientSet, i.Lister())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	pm := &v1alpha1.PrometheusServer{}
	if err := svc.EnsureCreation(ctx, pm); err != nil {
		t.Fatalf("unable to ensure service creation, error %v", err)
	}
	clActions := clientSet.Actions()
	if expected, got := 1, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	action := clActions[0]
	if expected, got := "create", action.GetVerb(); expected != got {
		t.Fatalf("unexpected verb, expected %s got %s", expected, got)
	}
	if expected, got := serviceResourceName, action.GetResource().Resource; expected != got {
		t.Fatalf("unexpected resource, expected %s got %s", expected, got)
	}
}

func TestItRemovesServiceOnDeletionRequest(t *testing.T) {
	clientSet := fake.NewSimpleClientset()

	sif := informers.NewSharedInformerFactory(clientSet, 0)
	i := sif.Core().V1().Services()

	svc := NewService(clientSet, i.Lister())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	pm := &v1alpha1.PrometheusServer{}
	if err := svc.EnsureDeletion(ctx, pm); err != nil {
		t.Fatalf("unable to ensure service deletion, error %v", err)
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
