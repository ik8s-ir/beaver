package main

import (
	"log"

	k8s "github.com/ik8s-ir/beaver/pkg/k8s"
	"github.com/ik8s-ir/beaver/pkg/ovsnet"
	"github.com/joho/godotenv"
	"k8s.io/client-go/tools/cache"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("The .env file not loaded")
	}
	ovsInformer := k8s.CreateOVSInformer()
	ovsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ovsnet.AddEvent,
	})
	stopCh := make(chan struct{})
	defer close(stopCh)
	go ovsInformer.Run(stopCh)

	select {}
}
