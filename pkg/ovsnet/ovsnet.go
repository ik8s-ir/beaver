package ovsnet

import (
	"log"

	"github.com/ik8s-ir/beaver/pkg/helpers"
	"github.com/ik8s-ir/beaver/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var converter = runtime.DefaultUnstructuredConverter
var bridge string = ""

func AddEvent(obj interface{}) {
	unstructuredObj := obj.(*unstructured.Unstructured)
	ovsnet := &types.OvsNet{}

	converter.FromUnstructured(unstructuredObj.Object, ovsnet)

	bridge = helpers.NextBridgeID(bridge, 4)
	log.Println(ovsnet.Namespace, ovsnet.Name)
	// ovsagent.CreateOVSNetwork(bridge, []string{"172.16.16.1"})
}
