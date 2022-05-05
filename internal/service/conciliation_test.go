package service

import (
	"context"
	"testing"

	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
)

func TestItRegisterHandlerAndMatchesOnConciliateState(t *testing.T) {
	c := NewConciliator()
	fh := &fakeConciliatorHandler{newState: v1alpha1.Running}
	c.Register(fh)

	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.Initializing

	newState, err := c.Conciliate(context.Background(), ps)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if expected, got := fh.newState, newState; expected != got {
		t.Errorf("new State does not match, expected %s got %s", expected, got)
	}
}

func TestItRegisterHandlerAndThrowsErrorOnNonMatcherFound(t *testing.T) {
	c := NewConciliator()
	fh := &fakeConciliatorHandler{newState: v1alpha1.Running}
	c.Register(fh)

	namespace := "default"
	name := "prometheus-server-crd"
	ps := getFakePrometheusServer(namespace, name)
	ps.Status.Phase = v1alpha1.Empty

	_, err := c.Conciliate(context.Background(), ps)
	if err == nil {
		t.Fatal("expected no handler error")
	}
}

type fakeConciliatorHandler struct {
	handled  int
	newState string
}

func (f *fakeConciliatorHandler) foo(ctx context.Context, ps *v1alpha1.PrometheusServer) (newStatus string, err error) {
	f.handled++
	return f.newState, nil
}

func (f *fakeConciliatorHandler) Handlers() map[string]StateHandler {
	return map[string]StateHandler{
		v1alpha1.Initializing: f.foo,
	}
}
