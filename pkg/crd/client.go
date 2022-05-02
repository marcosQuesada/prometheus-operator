package crd

import (
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// BuilBuildPrometheusServerInternalClient instantiates internal prometheus-server client
func BuilBuildPrometheusServerInternalClient() versioned.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("unable to get In cluster config, error %v", err)
	}

	client, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatalf("unable to build client from config, error %v", err)
	}

	return client
}

// BuildPrometheusServerExternalClient instantiates local prometheus-server client with local credentials
func BuildPrometheusServerExternalClient() versioned.Interface {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("unable to get cluster config from flags, error %v", err)
	}

	client, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatalf("unable to build client from config, error %v", err)
	}

	return client
}
