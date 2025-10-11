package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&Machine{}, &MachineList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty,omitzero"`

	Items []Machine `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// +required
	Spec MachineSpec `json:"spec"`
}

type MachineSpec struct {
	// +required
	// +kubebuilder:validation:Pattern="^([0-9a-f]{2}:){5}[0-9a-f]{2}$"
	MacAddress string `json:"macAddress"`

	// +required
	// +kubebuilder:validation:Pattern="^([1-9][0-9]{0,2}\\.){3}[1-9][0-9]{0,2}$"
	IPAddress string `json:"ipAddress"`

	// TODO: role, etc.
}
