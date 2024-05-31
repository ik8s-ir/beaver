package helpers

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
)

func CreateUniqueTimeName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
}

func NextBridgeID(c string, maxLen int) string {
	if c == "" {
		return "a"
	}
	if len(c) == maxLen && c == strings.Repeat("z", maxLen) {
		return ""
	}

	ca := []rune(c)
	i := len(ca) - 1
	for i >= 0 && ca[i] == 'z' {
		ca[i] = 'a'
		i--
	}
	if i < 0 {
		ca = append([]rune{'a'}, ca...)
	} else {
		ca[i]++
	}
	return string(ca)
}

func FindOtherNodesIpAddresses(nodes *v1.NodeList, nodeName string) []string {
	var result []string
	for _, node := range nodes.Items {
		if node.GetName() != nodeName {
			var internalIP string
			for _, address := range node.Status.Addresses {
				if address.Type == "InternalIP" {
					internalIP = address.Address
					break
				}
			}
			result = append(result, internalIP)
		}
	}
	return result
}
