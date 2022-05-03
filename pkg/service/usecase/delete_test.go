package usecase

import (
	"context"
	"errors"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"testing"
)

func TestItRemovesAllResourcesRemovesFinalizersAndJumpsToTerminatedStateOnSuccess(t *testing.T) {
	fn := &fakeFinalizer{}
	rm := &fakeResourceManager{}
	dl := NewDeleter(fn, rm).(*deleter)
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Finalizers = []string{v1alpha1.Name}

	newStatus, err := dl.Terminating(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on terminating state got %v", err)
	}

	if expected, got := v1alpha1.Terminated, newStatus; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}

	if expected, got := 1, rm.removeAll; expected != got {
		t.Errorf("total calls does not match, expected %d got %d", expected, got)
	}

	if expected, got := 1, fn.removeCalled; expected != got {
		t.Errorf("total calls does not match, expected %d got %d", expected, got)
	}
}

func TestItRemainsOnStateWhenRemovesAllWithResourceManagerError(t *testing.T) {
	fn := &fakeFinalizer{}
	rm := &fakeResourceManager{error: errors.New("foo error")}
	dl := NewDeleter(fn, rm).(*deleter)
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.Terminating
	ps.Finalizers = []string{v1alpha1.Name}

	newStatus, err := dl.Terminating(context.Background(), ps)
	if err == nil {
		t.Fatal("expected error")
	}

	if expected, got := ps.Status.Phase, newStatus; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}
}
