package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GroupName        = "k8slab.info"
	CrdKind   string = "PrometheusServer"
	Version   string = "v1alpha1"
	Singular  string = "prometheusserver"
	Plural    string = "prometheusservers"
	ShortName string = "pmts"
	Name      string = Plural + "." + GroupName
)

const (
	Empty           = ""
	Initializing    = "INITIALIZING"
	WaitingCreation = "WAITING_CREATION"
	Running         = "RUNNING"
	Reloading       = "RELOADING"
	WaitingRemoval  = "WAITING_REMOVAL"
	Terminating     = "TERMINATING"
	Terminated      = "TERMINATED"
)

// Status defines the observed state of Worker
type Status struct {
	Phase string `json:"phase,omitempty"`
}

// PrometheusServerSpec defines the desired state of PrometheusServer
type PrometheusServerSpec struct {
	Version string `json:"version"`
	Config  string `json:"config"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrometheusServer handles Prometheus Server stack.
type PrometheusServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrometheusServerSpec `json:"spec,omitempty"`
	Status Status               `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrometheusServerList contains a list of PrometheusServer
type PrometheusServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrometheusServer `json:"items"`
}

// @TODO: Segregate ¿=?¿?
func (in *PrometheusServer) HasFinalizer(s string) bool {
	if len(in.Finalizers) == 0 {
		return false
	}

	for _, v := range in.Finalizers {
		if v == s {
			return true
		}
	}
	return false
}
