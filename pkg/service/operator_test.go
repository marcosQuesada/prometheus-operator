package service

import (
	"context"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	crdFake "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned/fake"
	crdinformers "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stest "k8s.io/client-go/testing"
	"testing"
)

func TestITAddsFinalizerOnPrometheusServer(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	if ps.HasFinalizer(v1alpha1.Name) {
		t.Fatal("not expected to find finalizer")
	}

	pmClientSet := crdFake.NewSimpleClientset(ps)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	pi := crdInf.K8slab().V1alpha1().PrometheusServers()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go crdInf.Start(ctx.Done())

	r := []ResourceEnforcer{}
	o := NewOperator(pi.Lister(), pmClientSet, r)
	if err := o.addFinalizer(context.Background(), ps); err != nil {
		t.Fatalf("unable to add finalizer error %v", err)
	}

	if !ps.HasFinalizer(v1alpha1.Name) {
		t.Fatal("expected to find finalizer")
	}
}
func TestITRemovesFinalizerFromPrometheusServer(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	pmClientSet := crdFake.NewSimpleClientset(ps)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	pi := crdInf.K8slab().V1alpha1().PrometheusServers()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go crdInf.Start(ctx.Done())

	r := []ResourceEnforcer{}
	o := NewOperator(pi.Lister(), pmClientSet, r)
	_ = o.addFinalizer(context.Background(), ps)

	if err := o.removeFinalizer(context.Background(), ps); err != nil {
		t.Fatalf("unable to remove finalizer error %v", err)
	}

	if ps.HasFinalizer(v1alpha1.Name) {
		t.Fatal("not expected to find finalizer")
	}
}

func TestItCallsAllEnforcersOnDelete(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	fre := &FakeResourceEnforcer{}
	r := []ResourceEnforcer{fre}
	o := NewOperator(nil, nil, r)
	if err := o.delete(context.Background(), ps); err != nil {
		t.Fatalf("unapected error deleting resources, error %v", err)
	}

	if expected, got := 1, fre.deletion; expected != got {
		t.Fatalf("unexpected resource, expected %d got %d", expected, got)
	}
}

func TestRemoveFinalizer(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	pm := &v1alpha1.PrometheusServer{
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
	pmClientSet := crdFake.NewSimpleClientset(pm)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pi := crdInf.K8slab().V1alpha1().PrometheusServers()

	ps := getFakePrometheusServer(namespace, name)
	if err := pi.Informer().GetIndexer().Add(ps); err != nil {
		t.Fatalf("unable to add entry to indexer %v", err)
	}

	fre := &FakeResourceEnforcer{}
	r := []ResourceEnforcer{fre}
	o := NewOperator(pi.Lister(), pmClientSet, r)

	ps, err := pmClientSet.K8slabV1alpha1().PrometheusServers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unable to get Ps, error %v", err)
	}

	if err := o.removeFinalizer(ctx, ps); err != nil {
		t.Fatalf("unexpected error removing finalizer on %s error %v", name, err)
	}
	if ps.HasFinalizer(v1alpha1.Name) {
		t.Fatal("no finalizer expected")
	}
}

func TestItAddsFinalizerOnUpdatingPrometheusServerDefinition(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	pm := &v1alpha1.PrometheusServer{
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
	pmClientSet := crdFake.NewSimpleClientset(pm)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	pi := crdInf.K8slab().V1alpha1().PrometheusServers().Informer()

	if err := pi.GetIndexer().Add(pm); err != nil {
		t.Fatalf("unable to add crd to indexer, error %v", err)
	}

	r := []ResourceEnforcer{}
	o := NewOperator(crdInf.K8slab().V1alpha1().PrometheusServers().Lister(), pmClientSet, r)
	if err := o.Update(context.Background(), namespace, name); err != nil {
		t.Fatalf("unaexpected error updating, error %v", err)
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
	ps, ok := v.GetObject().(*v1alpha1.PrometheusServer)
	if !ok {
		t.Fatalf("unexpected type got %T", v.GetObject())
	}
	if !ps.HasFinalizer(v1alpha1.Name) {
		t.Fatal("expected to find finalizer")
	}
}

func TestItUpdatesStateFromInitializingToWaitingFiringResourceEnforcers(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	pm := &v1alpha1.PrometheusServer{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Finalizers: []string{v1alpha1.Name},
		},
		Spec: v1alpha1.PrometheusServerSpec{
			Version: "v0.0.1",
			Config:  "fakeConfigContent",
		},
		Status: v1alpha1.Status{Phase: v1alpha1.Initializing},
	}
	pmClientSet := crdFake.NewSimpleClientset(pm)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	pi := crdInf.K8slab().V1alpha1().PrometheusServers().Informer()

	if err := pi.GetIndexer().Add(pm); err != nil {
		t.Fatalf("unable to add crd to indexer, error %v", err)
	}

	fre := &FakeResourceEnforcer{}
	r := []ResourceEnforcer{fre}
	o := NewOperator(crdInf.K8slab().V1alpha1().PrometheusServers().Lister(), pmClientSet, r)
	if err := o.Update(context.Background(), namespace, name); err != nil {
		t.Fatalf("unaexpected error updating, error %v", err)
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
	if expected, got := 1, fre.creations; expected != got {
		t.Fatalf("unexpected resource, expected %d got %d", expected, got)
	}
	v, ok := action.(k8stest.UpdateAction)
	if !ok {
		t.Fatalf("unexpected type got %T", action)
	}
	ps, ok := v.GetObject().(*v1alpha1.PrometheusServer)
	if !ok {
		t.Fatalf("unexpected type got %T", v.GetObject())
	}
	if expected, got := v1alpha1.Waiting, ps.Status.Phase; expected != got {
		t.Fatalf("status does not match, expected %s got %s", expected, got)
	}
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

type FakeResourceEnforcer struct {
	creations int
	deletion  int
}

func (f *FakeResourceEnforcer) EnsureCreation(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	f.creations++
	return nil
}

func (f *FakeResourceEnforcer) EnsureDeletion(ctx context.Context, obj *v1alpha1.PrometheusServer) error {
	f.deletion++
	return nil
}

func (f *FakeResourceEnforcer) Name() string {
	return "fake"
}
