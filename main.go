package main

import (
	k8s "github.com/ik8s-ir/beaver/pkg/k8s"
)

func main() {

	go k8s.RunOVSInformer()

	select {}
}
