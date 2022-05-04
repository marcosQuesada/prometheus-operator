package operator

import (
	"context"
	"errors"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	crdFake "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned/fake"
	crdinformers "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestControllerItGetsCreatedOnListeningPodsWithPodAddition(t *testing.T) {
	namespace := "default"
	name := "foo"
	wg := &sync.WaitGroup{}
	wg.Add(1)
	eh := &fakeHandler{updateWaitGroup: wg}
	p := getFakePrometheusServer(namespace, name)
	pmClientSet := crdFake.NewSimpleClientset(p)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)
	pi := crdInf.K8slab().V1alpha1().PrometheusServers().Informer()
	ctl := NewController(eh, pi)
	ctl.workerFrequency = time.Millisecond * 50

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ctl.Run(ctx, 1)
	go crdInf.Start(ctx.Done())

	cache.WaitForCacheSync(ctx.Done(), pi.HasSynced)

	if err := waitUntilTimeout(wg, time.Millisecond*300); err != nil {
		t.Fatalf("wait error %v", err)
	}

	if expected, got := 1, eh.updated(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
	if expected, got := 0, eh.deleted(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}

	gvr := schema.GroupVersionResource{Resource: "prometheusservers"}
	gvk := schema.GroupVersionKind{Group: "k8slab.info", Version: "v1alpha1"}
	actions := []core.Action{
		core.NewListAction(gvr, gvk, namespace, metav1.ListOptions{}),
		core.NewWatchAction(gvr, namespace, metav1.ListOptions{}),
	}

	clActions := pmClientSet.Actions()
	if expected, got := 2, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	for i, action := range clActions {
		if len(actions) < i+1 {
			t.Errorf("%d unexpected actions: %+v", len(actions)-len(clActions), actions[i:])
			break
		}

		expectedAction := actions[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) && action.GetSubresource() == expectedAction.GetSubresource()) {
			t.Errorf("Expected %#v got %#v", expectedAction, action)
			continue
		}
	}

	if len(actions) > len(clActions) {
		t.Errorf("%d additional expected actions:%+v", len(actions)-len(clActions), actions[len(clActions):])
	}
}

func TestControllerItGetsDeletedOnListeningPodsWithPodAdditionWithoutBeingPreloadedInTheIndexer(t *testing.T) {
	namespace := "default"
	name := "foo"
	wg := &sync.WaitGroup{}
	wg.Add(1)
	eh := &fakeHandler{deleteWaitGroup: wg}

	pmClientSet := crdFake.NewSimpleClientset()
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)

	pi := crdInf.K8slab().V1alpha1().PrometheusServers().Informer()
	ctl := NewController(eh, pi)
	ctl.workerFrequency = time.Millisecond * 50

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ctl.Run(ctx, 1)
	go crdInf.Start(ctx.Done())

	cache.WaitForCacheSync(ctx.Done(), pi.HasSynced)

	p := getFakePrometheusServer(namespace, name)
	ctl.enqueuePrometheusServer(p)

	if err := waitUntilTimeout(wg, time.Millisecond*300); err != nil {
		t.Fatalf("wait error %v", err)
	}

	if expected, got := 0, eh.updated(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
	if expected, got := 1, eh.deleted(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
}

func TestControllerItRetriesConsumedEntriesOnHandlingErrorUntilMaxRetries(t *testing.T) {
	namespace := "default"
	name := "foo"
	eh := &fakeHandler{error: errors.New("foo error")}

	p := getFakePrometheusServer(namespace, name)
	pmClientSet := crdFake.NewSimpleClientset(p)
	crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, 0)

	pi := crdInf.K8slab().V1alpha1().PrometheusServers().Informer()
	ctl := NewController(eh, pi)
	ctl.workerFrequency = time.Millisecond * 50

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	go ctl.Run(ctx, 1)
	go crdInf.Start(ctx.Done())

	cache.WaitForCacheSync(ctx.Done(), pi.HasSynced)

	// informer runner needs time
	time.Sleep(time.Millisecond * 200)

	if expected, got := maxRetries+1, eh.updated(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
	if expected, got := 0, eh.deleted(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
}

type fakeHandler struct {
	totalUpdated    int32
	totalDeleted    int32
	error           error
	updateWaitGroup *sync.WaitGroup
	deleteWaitGroup *sync.WaitGroup
}

func (f *fakeHandler) Update(ctx context.Context, namespace, name string) error {
	atomic.AddInt32(&f.totalUpdated, 1)
	if f.updateWaitGroup != nil {
		f.updateWaitGroup.Done()
	}
	return f.error
}

func (f *fakeHandler) Delete(ctx context.Context, namespace, name string) error {
	atomic.AddInt32(&f.totalDeleted, 1)
	if f.deleteWaitGroup != nil {
		f.deleteWaitGroup.Done()
	}
	return f.error
}

func (f *fakeHandler) updated() int32 {
	return atomic.LoadInt32(&f.totalUpdated)
}

func (f *fakeHandler) deleted() int32 {
	return atomic.LoadInt32(&f.totalDeleted)
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

func waitUntilTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout error")
	}
}
