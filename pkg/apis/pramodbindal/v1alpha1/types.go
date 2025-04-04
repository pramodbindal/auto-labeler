package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:resource:path=labelers,scope=Namespaced,shortName=lbl,categories=all
// +genreconciler

type Labeler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LabelerSpec `json:"spec"`
}

type LabelerSpec struct {
	Labels         map[string]string `json:"labels"`
	Annotations    map[string]string `json:"annotations"`
	TargetResource string            `json:"targetResource"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type LabelerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Labeler `json:"items"`
}
