/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package betav1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SidecarconfSpec defines the desired state of Sidecarconf.
type SideCarContainerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// ContainerName is the name of the sidecar container.
	ContainerName string `json:"containerName,omitempty"`
	// ImageVersion is the version of the sidecar container image.
	ImageVersion string `json:"imageVersion,omitempty"`
	// Repo is the repository of the sidecar container image.
	Repo string `json:"repo,omitempty"`

	HeadlessServiceName string `json:"serviceName,omitempty"`
}

// SidecarconfStatus defines the observed state of Sidecarconf.
type SideCarContainerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SideCarContainer is the Schema for the ipconfs API.
type SideCarContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SideCarContainerSpec   `json:"spec,omitempty"`
	Status SideCarContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SidecarconfList contains a list of Sidecarconf.
type SideCarContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SideCarContainer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SideCarContainer{}, &SideCarContainerList{})
}
