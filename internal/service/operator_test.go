package service

import (
	"context"
	"testing"
	"time"

	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	crdFake "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned/fake"
	crdinformers "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stest "k8s.io/client-go/testing"
)

func TestItJumpsToNewStateWhenConciliationResultHasDifferentState(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	pm := getFakePrometheusServer(namespace, name)
	pmClientSet := crdFake.NewSimpleClientset(pm)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	pi := crdInf.K8slab().V1alpha1().PrometheusServers()

	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.WaitingCreation

	if err := pi.Informer().GetIndexer().Add(ps); err != nil {
		t.Fatalf("unable to add entry to indexer %v", err)
	}

	gc := &fakeCache{value: 1}
	c := &fakeConciliator{newState: v1alpha1.Running}
	o := NewOperator(pi.Lister(), pmClientSet, gc, c)

	if err := o.Update(context.Background(), namespace, name); err != nil {
		t.Fatalf("unexpected error updating, %v", err)
	}

	clActions := pmClientSet.Actions()
	if expected, got := 1, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	action := clActions[0]
	if expected, got := "update", action.GetVerb(); expected != got {
		t.Fatalf("unexpected verb, expected %s got %s", expected, got)
	}
	if expected, got := v1alpha1.Plural, action.GetResource().Resource; expected != got {
		t.Fatalf("unexpected resource, expected %s got %s", expected, got)
	}

	v, ok := action.(k8stest.UpdateAction)
	if !ok {
		t.Fatalf("unexpected type got %T", action)
	}
	ups, ok := v.GetObject().(*v1alpha1.PrometheusServer)
	if !ok {
		t.Fatalf("unexpected type got %T", v.GetObject())
	}

	if expected, got := v1alpha1.Running, ups.Status.Phase; expected != got {
		t.Fatalf("state does not match, expected %s got %s", expected, got)
	}
}

func TestItSetTerminatingStateOnPrometheusServerUpdateWithDeletionTimestamp(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	pm := getFakePrometheusServer(namespace, name)
	pmClientSet := crdFake.NewSimpleClientset(pm)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	pi := crdInf.K8slab().V1alpha1().PrometheusServers()

	ps := getFakePrometheusServer(namespace, name)
	now := metav1.NewTime(time.Now())
	ps.DeletionTimestamp = &now

	if err := pi.Informer().GetIndexer().Add(ps); err != nil {
		t.Fatalf("unable to add entry to indexer %v", err)
	}

	gc := &fakeCache{value: 1}
	c := &fakeConciliator{newState: v1alpha1.Running}
	o := NewOperator(pi.Lister(), pmClientSet, gc, c)

	if err := o.Update(context.Background(), namespace, name); err != nil {
		t.Fatalf("unexpected error updating, %v", err)
	}

	clActions := pmClientSet.Actions()
	if expected, got := 1, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	action := clActions[0]
	if expected, got := "update", action.GetVerb(); expected != got {
		t.Fatalf("unexpected verb, expected %s got %s", expected, got)
	}
	if expected, got := v1alpha1.Plural, action.GetResource().Resource; expected != got {
		t.Fatalf("unexpected resource, expected %s got %s", expected, got)
	}

	v, ok := action.(k8stest.UpdateAction)
	if !ok {
		t.Fatalf("unexpected type got %T", action)
	}
	ups, ok := v.GetObject().(*v1alpha1.PrometheusServer)
	if !ok {
		t.Fatalf("unexpected type got %T", v.GetObject())
	}

	if expected, got := v1alpha1.Terminating, ups.Status.Phase; expected != got {
		t.Fatalf("state does not match, expected %s got %s", expected, got)
	}
}

func TestItRemainsItsStateWhenConciliationReturnsSameCrdState(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	pm := getFakePrometheusServer(namespace, name)
	pmClientSet := crdFake.NewSimpleClientset(pm)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	pi := crdInf.K8slab().V1alpha1().PrometheusServers()

	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.Running

	if err := pi.Informer().GetIndexer().Add(ps); err != nil {
		t.Fatalf("unable to add entry to indexer %v", err)
	}

	gc := &fakeCache{value: 1}
	c := &fakeConciliator{newState: v1alpha1.Running}
	o := NewOperator(pi.Lister(), pmClientSet, gc, c)

	if err := o.Update(context.Background(), namespace, name); err != nil {
		t.Fatalf("unexpected error updating, %v", err)
	}

	clActions := pmClientSet.Actions()
	if expected, got := 0, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
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

type fakeConciliator struct {
	newState string
	error    error
}

func (f *fakeConciliator) Conciliate(ctx context.Context, ps *v1alpha1.PrometheusServer) (string, error) {
	return f.newState, f.error
}
