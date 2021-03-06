/*
Copyright 2021 Arata Furukawa.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +structType=atomic
type LocalSecretKeyReference struct {
	corev1.LocalObjectReference `json:",inline"`

	// +optional
	Key *string `json:"key,omitempty"`
}

type AccessTokenSource struct {
	// +optional
	SecretRef *LocalSecretKeyReference `json:"secretRef,omitempty"`
}

type SlackChannel struct {
	// The id of channel. e.g. C1234567890
	// +kubebuilder:validation:Pattern=`^C[0-9]+$`
	// +optional
	ID *string `json:"id,omitempty"`
	// The name of the channel. If an ID is specified, it will be ignored.
	// +optional
	Name *string `json:"name,omitempty"`
}

// SlackAppSpec represents information about an Slack App.
type SlackAppSpec struct {
	// +required
	AccessToken AccessTokenSource `json:"accessToken"`

	// +required
	// +kubebuilder:validation:MinItems=1
	Channels []SlackChannel `json:"channels"`
}

type PrivateKeySource struct {
	// +optional
	SecretRef *LocalSecretKeyReference `json:"secretRef,omitempty"`
}

// GitHubAppSpec represents information about an GitHub App.
type GitHubAppSpec struct {
	// +required
	AppId int64 `json:"appId"`

	// +required
	PrivateKey PrivateKeySource `json:"privateKey"`

	// +optional
	BaseURL *string `json:"baseURL,omitempty"`
}

// ProviderSpec defines the desired state of Provider
type ProviderSpec struct {
	// The type of this provider.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`
	// +kubebuilder:validation:MaxLength=64
	Type string `json:"type"`

	// +optional
	GitHubApp *GitHubAppSpec `json:"githubApp,omitempty"`
	// +optional
	SlackApp *SlackAppSpec `json:"slackApp,omitempty"`
}

// ProviderStatus defines the observed state of Provider
type ProviderStatus struct {
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Provider is the Schema for the providers API
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec   `json:"spec,omitempty"`
	Status ProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProviderList contains a list of Provider
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}
