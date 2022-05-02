package resource

import (
	"context"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	"k8s.io/client-go/informers"
	"testing"
	"time"
)

func TestFullStackCreation(t *testing.T) {
	clientSet := operator.BuildExternalClient()
	pmClientSet := crd.BuildPrometheusServerExternalClient()
	sif := informers.NewSharedInformerFactory(clientSet, 0)

	r := []operator.ResourceEnforcer{
		NewNamespace(clientSet, sif.Core().V1().Namespaces().Lister()),
		NewClusterRole(clientSet, sif.Rbac().V1().ClusterRoles().Lister()),
		NewClusterRoleBinding(clientSet, sif.Rbac().V1().ClusterRoleBindings().Lister()),
		NewConfigMap(clientSet, sif.Core().V1().ConfigMaps().Lister()),
		NewDeployment(clientSet, sif.Apps().V1().Deployments().Lister()),
		NewService(clientSet, sif.Core().V1().Services().Lister()),
	}
	op := operator.NewOperator(pmClientSet, r)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	p := &v1alpha1.PrometheusServer{

		Spec: v1alpha1.PrometheusServerSpec{
			Version: "0.0.1",
			Config:  cfg,
		},
	}
	if err := op.Update(ctx, nil, p); err != nil {
		t.Fatalf("unaexpected error updating, error %v", err)
	}
}

func TestFullStackDeletion(t *testing.T) {
	clientSet := operator.BuildExternalClient()
	pmClientSet := crd.BuildPrometheusServerExternalClient()
	sif := informers.NewSharedInformerFactory(clientSet, 0)

	r := []operator.ResourceEnforcer{
		NewNamespace(clientSet, sif.Core().V1().Namespaces().Lister()),
		NewClusterRole(clientSet, sif.Rbac().V1().ClusterRoles().Lister()),
		NewClusterRoleBinding(clientSet, sif.Rbac().V1().ClusterRoleBindings().Lister()),
		NewConfigMap(clientSet, sif.Core().V1().ConfigMaps().Lister()),
		NewDeployment(clientSet, sif.Apps().V1().Deployments().Lister()),
		NewService(clientSet, sif.Core().V1().Services().Lister()),
	}
	op := operator.NewOperator(pmClientSet, r)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	p := &v1alpha1.PrometheusServer{
		Spec: v1alpha1.PrometheusServerSpec{
			Version: "0.0.1",
			Config:  cfg,
		},
	}
	if err := op.Delete(ctx, p); err != nil {
		t.Fatalf("unaexpected error updating, error %v", err)
	}
}

//
//var cfg = `
//    global:
//      scrape_interval: 5s
//      evaluation_interval: 5s
//    rule_files:
//      - /etc/prometheus/prometheus.rules
//    alerting:
//      alertmanagers:
//      - scheme: http
//        static_configs:
//        - targets:
//          - "alertmanager.monitoring.svc:9093"
//
//    scrape_configs:
//      - job_name: 'node-exporter'
//        kubernetes_sd_configs:
//          - role: endpoints
//        relabel_configs:
//        - source_labels: [__meta_kubernetes_endpoints_name]
//          regex: 'node-exporter'
//          action: keep
//
//      - job_name: 'kubernetes-apiservers'
//
//        kubernetes_sd_configs:
//        - role: endpoints
//        scheme: https
//
//        tls_config:
//          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
//        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
//
//        relabel_configs:
//        - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
//          action: keep
//          regex: default;kubernetes;https
//
//      - job_name: 'kubernetes-nodes'
//
//        scheme: https
//
//        tls_config:
//          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
//        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
//
//        kubernetes_sd_configs:
//        - role: node
//
//        relabel_configs:
//        - action: labelmap
//          regex: __meta_kubernetes_node_label_(.+)
//        - target_label: __address__
//          replacement: kubernetes.default.svc:443
//        - source_labels: [__meta_kubernetes_node_name]
//          regex: (.+)
//          target_label: __metrics_path__
//          replacement: /api/v1/nodes/${1}/proxy/metrics
//
//      - job_name: 'kubernetes-pods'
//
//        kubernetes_sd_configs:
//        - role: pod
//
//        relabel_configs:
//        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
//          action: keep
//          regex: true
//        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
//          action: replace
//          target_label: __metrics_path__
//          regex: (.+)
//        - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
//          action: replace
//          regex: ([^:]+)(?::\d+)?;(\d+)
//          replacement: $1:$2
//          target_label: __address__
//        - action: labelmap
//          regex: __meta_kubernetes_pod_label_(.+)
//        - source_labels: [__meta_kubernetes_namespace]
//          action: replace
//          target_label: kubernetes_namespace
//        - source_labels: [__meta_kubernetes_pod_name]
//          action: replace
//          target_label: kubernetes_pod_name
//
//      - job_name: 'kube-state-metrics'
//        static_configs:
//          - targets: ['kube-state-metrics.kube-system.svc.cluster.local:8080']
//
//      - job_name: 'kubernetes-cadvisor'
//
//        scheme: https
//
//        tls_config:
//          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
//        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
//
//        kubernetes_sd_configs:
//        - role: node
//
//        relabel_configs:
//        - action: labelmap
//          regex: __meta_kubernetes_node_label_(.+)
//        - target_label: __address__
//          replacement: kubernetes.default.svc:443
//        - source_labels: [__meta_kubernetes_node_name]
//          regex: (.+)
//          target_label: __metrics_path__
//          replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor
//
//      - job_name: 'kubernetes-service-endpoints'
//
//        kubernetes_sd_configs:
//        - role: endpoints
//
//        relabel_configs:
//        - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
//          action: keep
//          regex: true
//        - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
//          action: replace
//          target_label: __scheme__
//          regex: (https?)
//        - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
//          action: replace
//          target_label: __metrics_path__
//          regex: (.+)
//        - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
//          action: replace
//          target_label: __address__
//          regex: ([^:]+)(?::\d+)?;(\d+)
//          replacement: $1:$2
//        - action: labelmap
//          regex: __meta_kubernetes_service_label_(.+)
//        - source_labels: [__meta_kubernetes_namespace]
//          action: replace
//          target_label: kubernetes_namespace
//        - source_labels: [__meta_kubernetes_service_name]
//          action: replace
//          target_label: kubernetes_name
//`

//func TestAddFinalizer(t *testing.T) {
//	namespace := "default"
//	name := "prometheus-server-crd"
//	p := &v1alpha1.PrometheusServer{
//		ObjectMeta: metav1.ObjectMeta{
//			Namespace: namespace,
//			Name: name,
//			Finalizers:
//		},
//		Spec: v1alpha1.PrometheusServerSpec{
//			Version: "0.0.1",
//			Config:  cfg,
//		},
//		Status:     v1alpha1.Status{},
//	}
//}
