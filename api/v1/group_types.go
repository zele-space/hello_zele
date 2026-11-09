package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GroupSpec defines the desired state of Group.
type GroupSpec struct {
	// Namespace is the target namespace to manage. If empty, the Group name is used.
	Namespace string `json:"namespace,omitempty"`
	// Quotas defines the ResourceQuota hard limits to apply in the namespace.
	Quotas map[string]string `json:"quotas,omitempty"`
}

// GroupStatus defines the observed state of Group.
type GroupStatus struct {
	// NamespaceCreated indicates whether the namespace exists.
	NamespaceCreated bool `json:"namespaceCreated,omitempty"`
	// ObservedNamespace is the namespace currently managed.
	ObservedNamespace string `json:"observedNamespace,omitempty"`
	// QuotaCreated indicates whether the ResourceQuota exists.
	QuotaCreated bool `json:"quotaCreated,omitempty"`
	// Conditions represent the latest available observations.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Group is the Schema for the groups API.
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec,omitempty"`
	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupList contains a list of Group.
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
}
