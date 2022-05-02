package resource

import (
	"context"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"testing"
	"time"
)

func TestServiceCreation(t *testing.T) {
	clientSet := operator.BuildExternalClient()

	sif := informers.NewSharedInformerFactory(clientSet, 0)
	i := sif.Core().V1().Services()

	svc := NewService(clientSet, i.Lister())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	pm := &v1alpha1.PrometheusServer{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.PrometheusServerSpec{},
		Status:     v1alpha1.Status{},
	}
	if err := svc.EnsureCreation(ctx, pm); err != nil {
		t.Fatalf("unable to ensure service creation, error %v", err)
	}
}

func TestServiceDeletion(t *testing.T) {
	clientSet := operator.BuildExternalClient()

	sif := informers.NewSharedInformerFactory(clientSet, 0)
	i := sif.Core().V1().Services()

	svc := NewService(clientSet, i.Lister())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	pm := &v1alpha1.PrometheusServer{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.PrometheusServerSpec{},
		Status:     v1alpha1.Status{},
	}
	if err := svc.EnsureDeletion(ctx, pm); err != nil {
		t.Fatalf("unable to ensure service deletion, error %v", err)
	}
}
