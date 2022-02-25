package main

import (
	"log"
	"time"

	c "github.com/db/snap/pkg/controller"
	snapinfofactorry "github.com/db/snap/pkg/generated/informers/externalversions"
	"github.com/db/snap/pkg/utils/client"
)

func main() {
	ch := make(chan struct{})
	defer close(ch)

	client, snapclient, exsnapclient := client.GetKubeClient()

	sharedInformers := snapinfofactorry.NewSharedInformerFactory(snapclient, 10*time.Minute)
	podInformer := sharedInformers.Snapshot().V1alpha1().Snapshots().Informer()

	controller := c.NewController(client, snapclient, exsnapclient, podInformer)
	controller.Run(ch)

	log.Println("Info: controller ", controller)

}
