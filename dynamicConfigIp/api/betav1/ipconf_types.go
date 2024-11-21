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

// IpconfSpec defines the desired state of Ipconf.
type IpconfSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//owner of the ipaddress
	Owner string `json:"owner,omitempty"`
	// Iface is interface that ipaddress configure on.
	Iface string `json:"iface,omitempty"`
	// Ipaddress .
	Ipaddress string `json:"ipaddress,omitempty"`
	// Netmask .
	Netmask string `json:"netmask,omitempty"`
}

// IpconfStatus defines the observed state of Ipconf.
type IpconfStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Ipconf is the Schema for the ipconfs API.
type Ipconf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpconfSpec   `json:"spec,omitempty"`
	Status IpconfStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IpconfList contains a list of Ipconf.
type IpconfList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ipconf `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ipconf{}, &IpconfList{})
}
