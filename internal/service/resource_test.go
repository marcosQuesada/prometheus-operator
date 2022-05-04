package service

import (
	"context"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestItCreatesAllResourcesOnEmptyCreation(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	fre := &fakeResourceEnforcer{}
	r := NewResource(fre)
	ps := getFakePrometheusServer(namespace, name)
	if err := r.CreateAll(context.Background(), ps); err != nil {
		t.Fatalf("unexpected error on resource creation got %v", err)
	}
	if expected, got := 1, fre.creations; expected != got {
		t.Fatalf("unexpected resource, expected %d got %d", expected, got)
	}
}

func TestItRecognizesAllResourcesCreated(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	fre := &fakeResourceEnforcer{exists: true}
	r := NewResource(fre)
	ps := getFakePrometheusServer(namespace, name)
	if err := r.CreateAll(context.Background(), ps); err != nil {
		t.Fatalf("unexpected error on resource creation got %v", err)
	}
	if expected, got := 1, fre.creations; expected != got {
		t.Fatalf("unexpected resource, expected %d got %d", expected, got)
	}

	res, err := r.AllCreated()
	if err != nil {
		t.Fatalf("unexpected error on checking all resource created, got %v", err)
	}
	if !res {
		t.Error("expected all resources created")
	}
}

func TestItRemovesAllResources(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	fre := &fakeResourceEnforcer{}
	r := NewResource(fre)
	ps := getFakePrometheusServer(namespace, name)
	if err := r.DeleteAll(context.Background(), ps); err != nil {
		t.Fatalf("unexpected error on resource creation got %v", err)
	}
	if expected, got := 1, fre.deletion; expected != got {
		t.Fatalf("unexpected resource, expected %d got %d", expected, got)
	}

	res, err := r.AllRemoved()
	if err != nil {
		t.Fatalf("unexpected error on checking all resource created, got %v", err)
	}
	if !res {
		t.Error("expected all resources deleted")
	}
}

type fakeResourceEnforcer struct {
	creations int
	deletion  int
	exists    bool
}

func (f *fakeResourceEnforcer) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	f.creations++
	return nil
}

func (f *fakeResourceEnforcer) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	f.deletion++
	return nil
}

func (f *fakeResourceEnforcer) IsCreated() (bool, error) {
	return f.exists, nil
}

func (f *fakeResourceEnforcer) Name() string {
	return "fake"
}
func getFakePrometheusServer(namespace, name string) *v1alpha1.PrometheusServer {
	return &v1alpha1.PrometheusServer{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.PrometheusServerSpec{
			Version: "v0.0.1",
			Config:  "fakeConfigContent",
		},
		Status: v1alpha1.Status{},
	}
}
