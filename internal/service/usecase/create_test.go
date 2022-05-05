package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestItAddsFinalizerAndStaysInTheSameStateOnEmptyState(t *testing.T) {
	fn := &fakeFinalizer{}
	rm := &fakeResourceManager{}
	c := NewCreator(fn, rm, &fakeRecorder{}).(*creator)
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	newStatus, err := c.Empty(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on empty state got %v", err)
	}

	if expected, got := ps.Status.Phase, newStatus; expected != got {
		t.Fatalf("unexpected status, expected %s got %s", expected, got)
	}

	if expected, got := 1, fn.addCalled; expected != got {
		t.Errorf("total calls does not match, expected %d got %d", expected, got)
	}
}

func TestItMovesToInitializingWhenAddedFinalizerOnEmptyState(t *testing.T) {
	fn := &fakeFinalizer{}
	rm := &fakeResourceManager{}
	c := NewCreator(fn, rm, &fakeRecorder{}).(*creator)
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Finalizers = []string{v1alpha1.Name}
	newStatus, err := c.Empty(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on empty state got %v", err)
	}

	if expected, got := v1alpha1.Initializing, newStatus; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}
}

func TestItRemainsOnSameStateOnErrorEnsuringFinalizerOnEmptyState(t *testing.T) {
	fn := &fakeFinalizer{error: errors.New("foo error")}
	rm := &fakeResourceManager{}
	c := NewCreator(fn, rm, &fakeRecorder{}).(*creator)
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	newStatus, err := c.Empty(context.Background(), ps)
	if err == nil {
		t.Fatal("expected error on empty state got")
	}

	if expected, got := v1alpha1.Empty, newStatus; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}
}

func TestItChecksAllResourcesAreCreatedOnWaitingCreationAndJumpsToRunningOnSuccess(t *testing.T) {
	fn := &fakeFinalizer{}
	rm := &fakeResourceManager{response: true}
	c := NewCreator(fn, rm, &fakeRecorder{}).(*creator)
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	newStatus, err := c.WaitingCreation(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on empty state got %v", err)
	}
	if expected, got := v1alpha1.Running, newStatus; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}
}

func TestItChecksAllResourcesAreCreatedOnWaitingCreationAndRemainsOnStateWhenAllResourcesStillPending(t *testing.T) {
	fn := &fakeFinalizer{}
	rm := &fakeResourceManager{response: false}
	c := NewCreator(fn, rm, &fakeRecorder{}).(*creator)
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.WaitingCreation

	newStatus, err := c.WaitingCreation(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on empty state got %v", err)
	}
	if expected, got := ps.Status.Phase, newStatus; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}
}

type fakeFinalizer struct {
	ensureCalled int
	addCalled    int
	removeCalled int
	error        error
}

func (f *fakeFinalizer) Ensure(ctx context.Context, ps *v1alpha1.PrometheusServer) error {
	f.ensureCalled++
	return f.error
}

func (f *fakeFinalizer) Add(ctx context.Context, ps *v1alpha1.PrometheusServer) error {
	f.addCalled++
	return f.error
}

func (f *fakeFinalizer) Remove(ctx context.Context, ps *v1alpha1.PrometheusServer) error {
	f.removeCalled++
	return f.error
}

type fakeResourceManager struct {
	removeAll int
	createAll int
	error     error
	response  bool
}

func (f *fakeResourceManager) AllCreated() (bool, error) {
	return f.response, f.error
}

func (f *fakeResourceManager) AllRemoved() (bool, error) {
	return f.response, f.error
}

func (f *fakeResourceManager) CreateAll(ctx context.Context, p *v1alpha1.PrometheusServer) error {
	f.createAll++
	return f.error
}

func (f *fakeResourceManager) DeleteAll(ctx context.Context, p *v1alpha1.PrometheusServer) error {
	f.removeAll++
	return f.error
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
