package service

import (
	"context"
	crdFake "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned/fake"
	"testing"
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

	if !HasFinalizer(ps) {
		t.Fatal("expected to find finalizer")
	}
}

func TestITRemovesFinalizerFromPrometheusServer(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)

	pmClientSet := crdFake.NewSimpleClientset(ps)
	o := NewFinalizer(pmClientSet)

	_ = o.Add(context.Background(), ps)

	if err := o.Remove(context.Background(), ps); err != nil {
		t.Fatalf("unable to remove finalizer error %v", err)
	}

	if HasFinalizer(ps) {
		t.Fatal("not expected to find finalizer")
	}
}
