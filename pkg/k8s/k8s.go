package k8s

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ik8s-ir/beaver/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var dynamicClient *dynamic.DynamicClient
var ovsInformer cache.SharedIndexInformer
var converter = runtime.DefaultUnstructuredConverter
var ovsnetResource = schema.GroupVersionResource{
	Group:    "networking.ik8s.ir",
	Version:  "v1alpha1",
	Resource: "ovsnets",
}

func CreateClient() *dynamic.DynamicClient {
	// singleton
	if dynamicClient != nil {
		return dynamicClient
	}

	config, err := createConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}

	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamicClient: %v", err)
	}
	return dynamicClient
}

func createConfig() (*rest.Config, error) {
	configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
	_, err := os.Stat(configFile)
	if err != nil {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", configFile)
}

func CreateOVSInformer() cache.SharedIndexInformer {
	if ovsInformer != nil {
		return ovsInformer
	}
	resource := schema.GroupVersionResource{Group: "networking.ik8s.ir", Version: "v1alpha1", Resource: "ovsnets"}
	informerfactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(CreateClient(), time.Second*30, "", nil)
	ovsInformer = informerfactory.ForResource(resource).Informer()
	return ovsInformer
}

func FetchComputeNodes() *v1.NodeList {
	nodeResource := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "nodes",
	}
	labelSelector := `node-role.kubernetes.io/compute`
	nodes := &v1.NodeList{}
	unstructuredNodes, err := dynamicClient.Resource(nodeResource).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		log.Fatalf("Error fetching compute nodes: %v", err)
	}
	converter.FromUnstructured(unstructuredNodes.UnstructuredContent(), nodes)
	return nodes
}

func GetOVSPodByNode(namespace string, nodeName string) *v1.Pod {
	podResource := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	pods := &v1.PodList{}
	unstructuredPods, err := dynamicClient.Resource(podResource).Namespace(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
		LabelSelector: "name=ik8s-ovs",
	})
	if err != nil {
		log.Printf("Error on fetching ovs pods list: %v \n", err)
		return nil
	}
	converter.FromUnstructured(unstructuredPods.UnstructuredContent(), pods)
	if len(pods.Items) > 0 {
		return &pods.Items[0]
	}
	return nil
}

func UpdateOVSnetBridge(on *types.OvsNet, bridge string) {
	on.Spec.Bridge = bridge

	unstructuredMap, _ := converter.ToUnstructured(on)
	unstructuredON := &unstructured.Unstructured{
		Object: unstructuredMap,
	}
	unstructuredON = addFinalizer(unstructuredON)
	_, err := dynamicClient.Resource(ovsnetResource).Namespace(on.GetNamespace()).Update(context.TODO(), unstructuredON, metav1.UpdateOptions{})
	if err != nil {
		log.Fatalf("error on updating ovsnet %s in namespace %s, error: %v \n", on.GetName(), on.GetNamespace(), err)
	}
}

func addFinalizer(resource *unstructured.Unstructured) *unstructured.Unstructured {
	finalizer := "finalizer.ovsnet.networking.ik8s.ir"
	finalizers := resource.GetFinalizers()
	for _, f := range finalizers {
		if f == finalizer {
			return resource
		}
	}
	resource.SetFinalizers(append(finalizers, finalizer))
	return resource
}

func DeleteFinalizers(resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	resource.SetFinalizers(nil)
	return dynamicClient.Resource(ovsnetResource).Namespace(resource.GetNamespace()).Update(context.TODO(), resource, metav1.UpdateOptions{})
}
