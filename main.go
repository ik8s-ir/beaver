package main

import (
	k8s "github.com/ik8s-ir/beaver/pkg/k8s"
	"github.com/ik8s-ir/beaver/pkg/ovsnet"
	"k8s.io/client-go/tools/cache"
)

func main() {

	ovsInformer := k8s.CreateOVSInformer()
	ovsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ovsnet.AddEvent,
	})
	stopCh := make(chan struct{})
	defer close(stopCh)
	go ovsInformer.Run(stopCh)

	select {}
}
