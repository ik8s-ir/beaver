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

type VswitchPostBody struct {
	Bridge   string         `json:"bridge"`
	Topology []MeshTopology `json:"topology"`
}

type MeshTopology struct {
	NodeIP string `json:"nodeIP"`
	VNI    int32  `json:"vni"`
}

type NetworkAttachmentDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NetworkAttachmentDefinitionSpec `json:"spec"`
}

type NetworkAttachmentDefinitionSpec struct {
	Config string `json:"config"`
}

type NetworkAttachmentDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NetworkAttachmentDefinition `json:"items"`
}
