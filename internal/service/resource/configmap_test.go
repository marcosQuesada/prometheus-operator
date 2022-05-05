package resource

import (
	"context"
	"testing"
	"time"

	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	k8stest "k8s.io/client-go/testing"
)

func TestItCreatesConfigMapOnCreationRequest(t *testing.T) {
	clientSet := fake.NewSimpleClientset()
	sif := informers.NewSharedInformerFactory(clientSet, 0)
	i := sif.Core().V1().ConfigMaps()

	svc := NewConfigMap(clientSet, i.Lister())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	version := "v1.0.1"
	fakeConfig := "fakeConfig"
	pm := &v1alpha1.PrometheusServer{
		Spec: v1alpha1.PrometheusServerSpec{Version: version, Config: fakeConfig},
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
	if expected, got := configMapResourceName, action.GetResource().Resource; expected != got {
		t.Fatalf("unexpected resource, expected %s got %s", expected, got)
	}

	v, ok := action.(k8stest.CreateAction)
	if !ok {
		t.Fatalf("unexpected type got %T", action)
	}
	d, ok := v.GetObject().(*v1.ConfigMap)
	if !ok {
		t.Fatalf("unexpected type got %T", v.GetObject())
	}

	cf, ok := d.Data[prometheusConfigMapKey]
	if !ok {
		t.Fatal("prometheus config not found in response")
	}

	if expected, got := fakeConfig, cf; expected != got {
		t.Fatalf("image version do not match, expected %s got %s", expected, got)
	}
}

func TestItDeletesConfigMapOnDeletionRequest(t *testing.T) {
	clientSet := fake.NewSimpleClientset()
	sif := informers.NewSharedInformerFactory(clientSet, 0)
	i := sif.Core().V1().ConfigMaps()

	svc := NewConfigMap(clientSet, i.Lister())
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
