package nad

import (
	"encoding/json"
	"log"

	"github.com/ik8s-ir/beaver/pkg/k8s"
	"github.com/ik8s-ir/beaver/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var converter = runtime.DefaultUnstructuredConverter

func AddEvent(obj interface{}) {
	unstructuredObj := obj.(*unstructured.Unstructured)
	nad := &types.NetworkAttachmentDefinition{}

	converter.FromUnstructured(unstructuredObj.Object, nad)

	log.Println(nad.GetNamespace(), nad.GetName())
	nadConfig := &types.NADConfig{}
	json.Unmarshal([]byte(nad.Spec.Config), nadConfig)
	log.Println(nadConfig.Type)
	if nad.GetName() == "default" || nadConfig.Type != "ovs" {
		return
	}
	k8s.CreateOVSnet(nad.GetName(), nad.GetNamespace())
	// k8s.DeleteOVSnet(nad.GetName(), nad.GetNamespace())
}
