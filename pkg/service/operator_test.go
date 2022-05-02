package service

import (
	"context"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestAddFinalizer(t *testing.T) {

}

func TestRemoveFinalizer(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	pmClientSet := crd.BuildPrometheusServerExternalClient()

	r := []ResourceEnforcer{}
	op := NewOperator(pmClientSet, r)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second+2)
	defer cancel()
	ps, err := pmClientSet.K8slabV1alpha1().PrometheusServers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unable to get Ps, error %v", err)
	}

	if err := op.removeFinalizer(ctx, ps); err != nil {
		t.Fatalf("unexpected error removing finalizer on %s error %v", name, err)
	}
}
