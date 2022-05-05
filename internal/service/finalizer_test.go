package service

import (
	"context"
	"testing"

	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	crdFake "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned/fake"
	k8stest "k8s.io/client-go/testing"
)

func TestITAddsFinalizerOnPrometheusServer(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	if HasFinalizer(ps) {
		t.Fatal("not expected to find finalizer")
	}

	pmClientSet := crdFake.NewSimpleClientset(ps)
	o := NewFinalizer(pmClientSet)
	if err := o.Add(context.Background(), ps); err != nil {
		t.Fatalf("unable to add finalizer error %v", err)
	}

	clActions := pmClientSet.Actions()
	if expected, got := 1, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	action := clActions[0]
	if expected, got := "update", action.GetVerb(); expected != got {
		t.Fatalf("unexpected verb, expected %s got %s", expected, got)
	}
	v, ok := action.(k8stest.UpdateAction)
	if !ok {
		t.Fatalf("unexpected type got %T", action)
	}

	p, ok := v.GetObject().(*v1alpha1.PrometheusServer)
	if !ok {
		t.Fatalf("unexpected type got %T", v.GetObject())
	}

	if !HasFinalizer(p) {
		t.Fatal("expected to find finalizer")
	}
}

func TestITRemovesFinalizerFromPrometheusServer(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Finalizers = []string{v1alpha1.Name}

	pmClientSet := crdFake.NewSimpleClientset(ps)
	o := NewFinalizer(pmClientSet)

	if err := o.Remove(context.Background(), ps); err != nil {
		t.Fatalf("unable to remove finalizer error %v", err)
	}
	clActions := pmClientSet.Actions()
	if expected, got := 1, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	action := clActions[0]
	if expected, got := "update", action.GetVerb(); expected != got {
		t.Fatalf("unexpected verb, expected %s got %s", expected, got)
	}
	v, ok := action.(k8stest.UpdateAction)
	if !ok {
		t.Fatalf("unexpected type got %T", action)
	}

	p, ok := v.GetObject().(*v1alpha1.PrometheusServer)
	if !ok {
		t.Fatalf("unexpected type got %T", v.GetObject())
	}

	if HasFinalizer(p) {
		t.Fatal("not expected to find finalizer")
	}
}
