package nad

import (
	"encoding/json"
	"log"

	"github.com/ik8s-ir/beaver/pkg/helpers"
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

	var nadConfig map[string]interface{}
	err := json.Unmarshal([]byte(nad.Spec.Config), &nadConfig)
	if err != nil {
		log.Printf("Error unmarshaling nad config json: %v \n", err)
		return
	}

	if nad.GetName() == "default" || nadConfig["type"] != "ovs" {
		return
	}
	if nadConfig["bridge"] != "" {
		return
	}

	nad.SetFinalizers([]string{"ovsnet.networking.ik8s.ir"})
	for {
		bridgeName := helpers.CreateUniqueTimeName()
		_, err := k8s.CreateOVSnet(bridgeName, nad.GetNamespace(), nad.GetName())
		if err != nil {
			log.Printf("error on creating %s (cluster level) for namespace %s, err: %v \n retrying...\n", nad.GetName(), nad.GetNamespace(), err)
		} else {
			nadConfig["bridge"] = bridgeName
			jsonNadConfig, _ := json.Marshal(nadConfig)
			nad.Spec.Config = string(jsonNadConfig)
			nadMap, _ := converter.ToUnstructured(nad)
			for {
				_, err := k8s.UpdateNAD(&unstructured.Unstructured{Object: nadMap})
				if err == nil {
					break
				}
				log.Printf("Error updating NAD %s at the namespace %s with ovs bridge named: %s, error: %v \nretrying... \n", nad.GetName(), nad.GetNamespace(), bridgeName, err)
			}
			break
		}
	}
	// k8s.DeleteOVSnet(nad.GetName())
}

func UpdateEvent(_, obj interface{}) {
	unstructuredObj := obj.(*unstructured.Unstructured)
	nad := &types.NetworkAttachmentDefinition{}
	converter.FromUnstructured(unstructuredObj.Object, nad)

	var nadConfig map[string]interface{}
	err := json.Unmarshal([]byte(nad.Spec.Config), &nadConfig)
	if err != nil {
		log.Printf("Error unmarshaling nad config json: %v \n", err)
		return
	}

	if nad.GetName() == "default" || nadConfig["type"] != "ovs" || nad.GetDeletionTimestamp() == nil {
		return
	}

	bridgeName := nadConfig["bridge"].(string)

	for {
		err := k8s.DeleteOVSnet(bridgeName)
		if err != nil {
			log.Printf("error on deleting ovsnet %s for nad %s (cluster level) for namespace %s, err: %v \n retrying...\n", bridgeName, nad.GetName(), nad.GetNamespace(), err)
			continue
		}
		_, err = k8s.DeleteNADFinalizers(unstructuredObj)
		if err != nil {
			log.Printf("error on deleting finalizers for nad %s in namespaces %s", nad.GetName(), nad.GetNamespace())
			continue
		}
		break
	}
}
