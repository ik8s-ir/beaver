package main

import (
	k8s "github.com/ik8s-ir/beaver/pkg/k8s"
)

func main() {

	ovsInformer := k8s.RunOVSInformer()
	stopCh := make(chan struct{})
	defer close(stopCh)
	go ovsInformer.Run(stopCh)

	select {}
}
