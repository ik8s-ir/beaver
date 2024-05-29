package k8s

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ik8s-ir/beaver/pkg/ovsnet"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset *dynamic.DynamicClient

func CreateClient() *dynamic.DynamicClient {
	// singleton
	if clientset != nil {
		return clientset
	}

	config, err := createConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}

	clientset, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}
	return clientset
}

func createConfig() (*rest.Config, error) {
	configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
	_, err := os.Stat(configFile)
	if err != nil {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", configFile)
}

func RunOVSInformer() cache.SharedIndexInformer {
	resource := schema.GroupVersionResource{Group: "networking.ik8s.ir", Version: "v1alpha1", Resource: "ovsnets"}
	informerfactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(CreateClient(), time.Second*30, "", nil)
	ovsInformer := informerfactory.ForResource(resource).Informer()
	ovsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ovsnet.AddEvent,
	})
	return ovsInformer
}
