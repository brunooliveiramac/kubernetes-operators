/*
Copyright 2021.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Stack struct {
	Provider string   `json:"provider"`
	Registry string   `json:"registry"`
	Image    string   `json:"image"`
	Version  string   `json:"version"`
	Region   string   `json:"region"`
	Outputs  []string `json:"outputs"`
}

// StackInstanceSpec defines the desired state of StackInstance
type StackInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of StackInstance. Edit stackinstance_types.go to remove/update
	Replicas int32 `json:"replicas,omitempty"`

	Stack Stack `json:"stack"`
}

// StackInstanceStatus defines the observed state of StackInstance
type StackInstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PodNames          []string `json:"podNames"`
	AvailableReplicas int32    `json:"availableReplicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StackInstance is the Schema for the stackinstances API
type StackInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackInstanceSpec   `json:"spec,omitempty"`
	Status StackInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StackInstanceList contains a list of StackInstance
type StackInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StackInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StackInstance{}, &StackInstanceList{})
}
