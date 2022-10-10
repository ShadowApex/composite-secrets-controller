/*
Copyright 2022.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectRef defines a reference to a Secret or ConfigMap
type ObjectRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

// Replacement defines the source we will replace a value with
type Replacement struct {
	SecretRef    *ObjectRef `json:"secretRef,omitempty"`
	ConfigMapRef *ObjectRef `json:"configMapRef,omitempty"`
}

// Template defines a secret template. The data or stringData can use replacement
// strings in its contents which will be replaced by values defined in the
// 'replacements' section.
type Template struct {
	// Immutable, if set to true, ensures that data stored in the Secret cannot
	// be updated (only object metadata can be modified).
	// If not set to true, the field can be modified at any time.
	// Defaulted to nil.
	// +optional
	Immutable *bool `json:"immutable,omitempty" protobuf:"varint,5,opt,name=immutable"`

	// stringData allows specifying non-binary secret data in string form.
	// It is provided as a write-only input field for convenience.
	// All keys and values are merged into the data field on write, overwriting any existing values.
	// The stringData field is never output when reading from the API.
	// +k8s:conversion-gen=false
	// +optional
	StringData map[string]string `json:"stringData" protobuf:"bytes,4,rep,name=stringData"`

	// Used to facilitate programmatic handling of secret data.
	// More info: https://kubernetes.io/docs/concepts/configuration/secret/#secret-types
	// +optional
	Type v1.SecretType `json:"type,omitempty" protobuf:"bytes,3,opt,name=type,casttype=SecretType"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CompositeSecretSpec defines the desired state of CompositeSecret
type CompositeSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replacements maps a string to replace in the template with a value
	// in a configmap or secret
	Replacements map[string]*Replacement `json:"replacements,omitempty"`
	// Template defines the secret template to replace values from other
	// sources
	Template *Template `json:"template"`
}

// CompositeSecretStatus defines the observed state of CompositeSecret
type CompositeSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Synced is the status of generating a secret based on the given
	// replacements
	Synced bool `json:"synced"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CompositeSecret is the Schema for the compositesecrets API
type CompositeSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CompositeSecretSpec   `json:"spec,omitempty"`
	Status CompositeSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CompositeSecretList contains a list of CompositeSecret
type CompositeSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CompositeSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CompositeSecret{}, &CompositeSecretList{})
}
