package ovsnet

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ik8s-ir/beaver/pkg/helpers"
	"github.com/ik8s-ir/beaver/pkg/k8s"
	"github.com/ik8s-ir/beaver/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var converter = runtime.DefaultUnstructuredConverter
var bridge string = ""

func AddEvent(obj interface{}) {
	unstructuredObj := obj.(*unstructured.Unstructured)
	on := &types.OvsNet{}

	converter.FromUnstructured(unstructuredObj.Object, on)

	bridge = helpers.NextBridgeID(bridge, 12)
	log.Println(on.Namespace, on.Name)
	CreateDestributedVswitch(bridge)
	// ovsagent.CreateOVSNetwork(bridge, []string{"172.16.16.1"})

	// 1. retrieve all kubernetes compute nodes.
}

func CreateDestributedVswitch(bridge string) {
	// 1. fertch compute nodes
	nodes := k8s.FetchComputeNodes()

	for _, node := range nodes.Items {
		pod := k8s.GetOVSPodByNode("kube-system", node.GetName())
		done := false
		retry := 0
		for !done {
			_, err := createVswitch(bridge, "http://"+pod.GetName()+".kube-system.svc.cluster.local")
			if err == nil {
				done = true
			}
			log.Printf("Failed to create vswitch on pod %s , node %s.\n %v \n", pod.GetName(), node.GetName(), err)

			retry++
			if retry > 10 {
				log.Fatalf("Hasn't success after 10 times.")
			}
			log.Printf("Retry %v/10 in 2 seconds ...\n", retry)
			time.Sleep(2 * time.Second)
		}
	}
}

func createVswitch(bridge string, url string) (resp *http.Response, err error) {
	body := &types.VswitchPostBody{
		Bridge: bridge,
	}
	jsonBody, _ := json.Marshal(body)
	return http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}
