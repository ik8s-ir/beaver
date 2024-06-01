package ovsnet

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ik8s-ir/beaver/pkg/helpers"
	"github.com/ik8s-ir/beaver/pkg/k8s"
	"github.com/ik8s-ir/beaver/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var now = time.Now()
var converter = runtime.DefaultUnstructuredConverter
var lastVNI int32 = 100

func AddEvent(obj interface{}) {
	unstructuredObj := obj.(*unstructured.Unstructured)
	if unstructuredObj.GetCreationTimestamp().Time.Before(now) {
		return
	}

	on := &types.OvsNet{}
	converter.FromUnstructured(unstructuredObj.Object, on)

	bridge := on.GetName()

	CreateDestributedVswitch(bridge)
	// k8s.UpdateOVSnetBridge(on, bridge)
}

func UpdateEvent(_, obj interface{}) {
	unstructuredObj := obj.(*unstructured.Unstructured)
	on := &types.OvsNet{}
	converter.FromUnstructured(unstructuredObj.Object, on)
	log.Printf("updating... %s %s", on.GetName(), on.GetDeletionTimestamp())
	if on.GetDeletionTimestamp() != nil {
		DeleteDestributedVswitch(on.GetName())
		_, err := k8s.DeleteOVSNetFinalizers(unstructuredObj)
		if err != nil {
			log.Printf("finalizers deletion was failed on %s at the namespace %s.\n error: %v \n", on.GetName(), on.GetNamespace(), err)
		}
	}
}

func CreateDestributedVswitch(bridge string) {
	// 1. fertch compute nodes
	nodes := k8s.FetchComputeNodes()
	for _, node := range nodes.Items {
		nodesIPs := helpers.FindOtherNodesIpAddresses(nodes, node.GetName())
		retry := 0
		for {
			pod := k8s.GetOVSPodByNode(os.Getenv("NAMESPACE"), node.GetName())
			if pod == nil {
				log.Printf("There's no ik8s-ovs pod on node %s", node.GetName())
				break
			}
			var url string
			if os.Getenv("ENV") == "development" {
				url = "http://172.16.220.10:8000/v1alpha1/ovs"
			} else {
				url = "http://" + pod.GetName() + os.Getenv("NAMESPACE") + ".svc.cluster.local:8000/ovs"
			}
			res, err := createVswitch(bridge, url, nodesIPs)
			if err != nil && res == nil {
				log.Println(err)
				break
			}
			if err == nil && res.StatusCode == http.StatusOK {
				break
			}

			body, _ := io.ReadAll(res.Body)
			log.Printf("Received %v, Failed to create vswitch on pod %s , node %s.\n", res.StatusCode, pod.GetName(), node.GetName())
			log.Printf("error: %v Response: %s", err, string(body))

			retry++
			if retry > 10 {
				log.Println("Hasn't success after 10 times.")
				break
			}
			log.Printf("Retry %v/10 \n", retry)
		}
	}
	lastVNI += int32(len(nodes.Items))
}

func DeleteDestributedVswitch(bridge string) {
	log.Printf("deleting %s ...", bridge)
	nodes := k8s.FetchComputeNodes()
	// var failedNodes []string
	for _, node := range nodes.Items {
		retry := 0
		for {
			pod := k8s.GetOVSPodByNode(os.Getenv("NAMESPACE"), node.GetName())
			if pod == nil {
				log.Printf("There's no ik8s-ovs pod on node %s namespace %s", node.GetName(), os.Getenv("NAMESPACE"))
				// failedNodes = append(failedNodes, node.GetName())
				continue
			}
			var url string
			if os.Getenv("ENV") == "development" {
				url = "http://172.16.220.10:8000/v1alpha1/ovs"
			} else {
				url = "http://" + pod.GetName() + os.Getenv("NAMESPACE") + ".svc.cluster.local:8000/ovs"
			}
			res, err := deleteVswitch(bridge, url)
			if err != nil && res == nil {
				log.Println(err)
				break
			}
			if err == nil && res.StatusCode == http.StatusOK {
				break
			}

			body, _ := io.ReadAll(res.Body)
			log.Printf("Received %v, Failed to delete vswitch on pod %s , node %s.\n", res.StatusCode, pod.GetName(), node.GetName())
			log.Printf("error: %v Response: %s", err, string(body))

			retry++
			if retry > 10 {
				log.Printf("Hasn't success after 10 times.")
				break
			}
			log.Printf("Retry %v/10 ...\n", retry)
		}
	}
}

func createVswitch(bridge string, url string, nodeIps []string) (resp *http.Response, err error) {
	var topology []types.MeshTopology
	vni := lastVNI
	for _, nodeIP := range nodeIps {
		vni++
		topology = append(topology, types.MeshTopology{
			NodeIP: nodeIP,
			VNI:    vni,
		})
	}

	body := &types.VswitchPostBody{
		Bridge:   bridge,
		Topology: topology,
	}
	jsonBody, _ := json.Marshal(body)
	return http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}

func deleteVswitch(bridge string, url string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodDelete, url+"/"+bridge, nil)
	return client.Do(req)
}
