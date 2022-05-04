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
	// Empty happens on PrometheusServer crd creation
	Empty = ""
	// Initializing happens on resource creation
	Initializing = "INITIALIZING"
	// WaitingCreation happens waiting all resources created
	WaitingCreation = "WAITING_CREATION"
	// Running is the conciliation target
	Running = "RUNNING"
	// Reloading happens when received PrometheusServer update while running
	Reloading = "RELOADING"
	// WaitingRemoval happens on reloading first stage, waits until all resources are removed
	WaitingRemoval = "WAITING_REMOVAL"
	// Terminating happens on PrometheusServer marked to delete
	Terminating = "TERMINATING"
	// Terminated happens after processing Terminate, final exit state
	Terminated = "TERMINATED"
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
