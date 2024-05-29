package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OvsNet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              OvsNetSpec `json:"spec"`
}

type OvsNetSpec struct {
	Bridge string `json:"bridge"`
	Status string `jsob:"status"`
}
