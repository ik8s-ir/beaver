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
	if on.GetDeletionTimestamp() != nil {
		log.Printf("updating... %s %s", on.GetName(), on.GetDeletionTimestamp())
		DeleteDestributedVswitch(on.GetName())
		_, err := k8s.DeleteOVSNetFinalizers(unstructuredObj)
		if err != nil {
			log.Printf("ovsnet finalizers deletion was failed on %s .\n error: %v \n", on.GetName(), err)
		}
	}
}

func CreateDestributedVswitch(bridge string) {
	lastVNI := findLastVNI()
	nodes := k8s.FetchComputeNodes()
	for _, node := range nodes.Items {
		lastVNI := findLastVNI()
		nodesIPs := helpers.FindOtherNodesIpAddresses(nodes, node.GetName())
		url := createURL(helpers.FindNodeInternalIPAddress(node))
		retry := 0
		for {

			pod := k8s.GetOVSPodByNode(os.Getenv("NAMESPACE"), node.GetName())
			if pod == nil {
				log.Printf("There's no ik8s-ovs pod on node %s", node.GetName())
				break
			}

			res, err := createVswitch(bridge, url, nodesIPs, lastVNI)
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
	lastVNI += len(nodes.Items)
	updateLastVNI(lastVNI)
}

func DeleteDestributedVswitch(bridge string) {
	log.Printf("deleting %s ...", bridge)
	nodes := k8s.FetchComputeNodes()

	// var failedNodes []string
	for _, node := range nodes.Items {
		url := createURL(helpers.FindNodeInternalIPAddress(node))
		retry := 0
		for {
			pod := k8s.GetOVSPodByNode(os.Getenv("NAMESPACE"), node.GetName())
			if pod == nil {
				log.Printf("There's no ik8s-ovs pod on node %s namespace %s", node.GetName(), os.Getenv("NAMESPACE"))
				// failedNodes = append(failedNodes, node.GetName())
				continue
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

func createVswitch(bridge string, url string, nodeIps []string, lastVNI int) (resp *http.Response, err error) {
	var topology []types.MeshTopology
	for _, nodeIP := range nodeIps {
		vni := lastVNI + 1
		topology = append(topology, types.MeshTopology{
			NodeIP: nodeIP,
			VNI:    int32(vni),
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

func createURL(ip string) string {
	var url string
	if os.Getenv("ENV") == "development" {
		url = "http://172.16.220.10:8000/v1alpha1/ovs"
	} else {
		url = "http://" + ip + ":8000/v1alpha1/ovs"
	}
	return url
}

func findLastVNI() int {
	for {
		c := 0
		v, err := k8s.GetLastOVSVNI()
		if err != nil {
			log.Printf("try %d, Error on getting the last ovs vni: %v \n", c, err)
			time.Sleep(time.Second * 1)
			c++
			continue
		}
		if v == nil {
			_, err := k8s.CreateOVSVNI("last", 200)
			if err != nil {
				log.Fatalf("error on creating initial OVS VNI: %v", err)
			}
			return 200
		}
		vni := &types.OVSVNI{}
		converter.FromUnstructured(v.Object, vni)
		labels := vni.GetLabels()
		if labels["mutext"] != "" {
			log.Printf("The last vni are in use, wait till mutex remove")
			continue
		}

		return vni.Spec.VNI
	}
}

func updateLastVNI(vni int) {
	c := 0
	for {
		lastVNIunstructured, err := k8s.GetLastOVSVNI()
		if err != nil {
			log.Printf("try %d, Error on getting the last ovs vni: %v \n", c, err)
			time.Sleep(time.Second * 1)
			c++
			continue
		}
		lastVNI := &types.OVSVNI{}
		converter.FromUnstructured(lastVNIunstructured.Object, lastVNI)
		labels := map[string]string{
			"mutext": "",
		}
		lastVNI.SetLabels(labels)
		lastVNI.Spec.VNI = vni
		unstructuredMap, _ := converter.ToUnstructured(lastVNI)
		unstructuredON := &unstructured.Unstructured{
			Object: unstructuredMap,
		}
		_, err = k8s.UpdateOVSVNI(unstructuredON)
		if err != nil {
			log.Printf("try %d, Error on updating the last ovs vni: %v \n", c, err)
			continue
		}
		break
	}
}
