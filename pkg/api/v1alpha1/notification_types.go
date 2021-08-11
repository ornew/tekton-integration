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

type TaskRunFilter struct {
	// +required
	Enabled bool `json:"enabled"`
}

type PipelineRunFilter struct {
	// +required
	Enabled bool `json:"enabled"`
}

// RunFilter defines rules for filtering Tekton Run objects.
type RunFilter struct {
	// +optional
	TaskRun *TaskRunFilter `json:"taskRun,omitempty"`

	// +optional
	PipelineRun *PipelineRunFilter `json:"pipelineRun,omitempty"`

	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

// NotificationSpec defines the desired state of Notification
type NotificationSpec struct {
	// Handle events using this provider.
	// +required
	ProviderRef corev1.LocalObjectReference `json:"providerRef"`

	// This flag tells the controller to suspend subsequent events dispatching.
	// Defaults to false.
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

// NotificationStatus defines the observed state of Notification
type NotificationStatus struct {
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Notification is the Schema for the notifications API
type Notification struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotificationSpec   `json:"spec,omitempty"`
	Status NotificationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NotificationList contains a list of Notification
type NotificationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Notification `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Notification{}, &NotificationList{})
}
