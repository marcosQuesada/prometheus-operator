package usecase

import (
	"context"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestItRemainsRunningOnUpdateWithSameGeneration(t *testing.T) {
	c := &fakeCache{value: 1}
	rm := &fakeResourceManager{}
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.Running
	ps.Generation = 1
	r := NewReloader(c, rm, &fakeRecorder{}).(*reloader)
	newState, err := r.Running(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on terminating state got %v", err)
	}

	if expected, got := ps.Status.Phase, newState; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}
}

func TestItStartsReloadingOnUpdateWithNewerGeneration(t *testing.T) {
	c := &fakeCache{value: 1}
	rm := &fakeResourceManager{}
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.Running
	ps.Generation = 2
	r := NewReloader(c, rm, &fakeRecorder{}).(*reloader)
	newState, err := r.Running(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on terminating state got %v", err)
	}

	if expected, got := v1alpha1.Reloading, newState; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}
}

func TestItRemovesAllResourcesAndJumpsToWaitingRemovalState(t *testing.T) {
	c := &fakeCache{value: 1}
	rm := &fakeResourceManager{}
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.Reloading

	r := NewReloader(c, rm, &fakeRecorder{}).(*reloader)
	newState, err := r.Reloading(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on terminating state got %v", err)
	}

	if expected, got := v1alpha1.WaitingRemoval, newState; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}

	if expected, got := 1, rm.removeAll; expected != got {
		t.Errorf("total calls does not match, expected %d got %d", expected, got)
	}
}

func TestItChecksAllResourcesAreRemovedAndJumpsToInitializeToForceResourceRecreation(t *testing.T) {
	c := &fakeCache{value: 1}
	rm := &fakeResourceManager{response: true}
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.WaitingRemoval

	r := NewReloader(c, rm, &fakeRecorder{}).(*reloader)
	newState, err := r.WaitingRemoval(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error on terminating state got %v", err)
	}

	if expected, got := v1alpha1.Initializing, newState; expected != got {
		t.Fatalf("new state does not match, expected %s got %s", expected, got)
	}
}

type fakeCache struct {
	value   int64
	set     int
	removed int
}

func (f *fakeCache) Get(namespace, name string) int64 {
	return f.value
}

func (f *fakeCache) Set(namespace, name string, v int64) {
	f.set++
}

func (f *fakeCache) Remove(namespace, name string) {
	f.removed++
}

type fakeRecorder struct{}

func (f *fakeRecorder) Event(object runtime.Object, eventtype, reason, message string) {}

func (f *fakeRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}

func (f *fakeRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}
