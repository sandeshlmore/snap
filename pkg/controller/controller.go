package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	snapclientset "github.com/db/snap/pkg/generated/clientset/versioned"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	exsnapclientset "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Controller struct {
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface

	snapclient   snapclientset.Interface
	snapinformer cache.SharedIndexInformer
	exsnapclient exsnapclientset.Interface
}

func NewController(clientset kubernetes.Interface, snapclient snapclientset.Interface,
	exsnapclient exsnapclientset.Interface, snapinformer cache.SharedIndexInformer) *Controller {

	c := &Controller{
		clientset:    clientset,
		queue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "myqueue"),
		snapinformer: snapinformer,
		snapclient:   snapclient,
		exsnapclient: exsnapclient,
	}

	snapinformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.HandleAddEvent,
		},
	)
	return c
}

func (c *Controller) Run(ch <-chan struct{}) {
	log.Println("Info: starting controller")

	defer c.queue.ShutDown()

	go c.snapinformer.Run(ch)
	if !cache.WaitForCacheSync(ch, c.snapinformer.HasSynced) {
		log.Print("Info: waiting for cache to be synced\n")
	}

	go wait.Until(c.worker, 1*time.Second, ch)
	<-ch
}

func (c *Controller) worker() {
	for c.processItem() {

	}
}

func (c *Controller) processItem() bool {
	log.Println("Info: In ProcessItem")
	item, shutdown := c.queue.Get()
	defer c.queue.Forget(item)

	if shutdown {
		return false
	}

	key, err := cache.MetaNamespaceKeyFunc(item)

	if err != nil {
		log.Printf("Error: getting key from cache %s\n", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		log.Printf("Error: splitting key into namespace and name %s\n", err.Error())
		return false
	}

	log.Println("Info: In namespace", ns, " pvc name ", name)

	snap, error := c.snapclient.SnapshotV1alpha1().Snapshots(ns).Get(context.Background(), name, metav1.GetOptions{})
	if error != nil {
		log.Printf("Error: can not find snapshot %s in namespace %s\n", name, ns)
		log.Println(err)
	}
	log.Println("Info: After getting snap", snap.Spec.SnapShotPVCSelector)

	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", "snapshotpvcselector", snap.Spec.SnapShotPVCSelector),
	}

	pvcList, err := c.clientset.CoreV1().PersistentVolumeClaims("").List(context.Background(), filters)
	if err != nil {
		log.Println("Error: can not list PersistentVolumeClaims")
		log.Println("Error: ", err)
	}
	log.Println("Info: After getting pvclist")

	for _, pvc := range pvcList.Items {
		log.Printf("\nInfo: Taking Snapshot for PVC %s\n", pvc.Name)
		err = CreateSnapshot(pvc, c)
		if err != nil {
			log.Printf("Error: Failed to create snapshot for pvc %s\n", pvc.Name)
			log.Println("Error: ", err)
		}
	}
	log.Println("Info: Done Processing In namespace", ns, " pvc name ", name)
	return true
}

// Create New snapshot for a PVC
func CreateSnapshot(pvc corev1.PersistentVolumeClaim, c *Controller) error {
	currentTime := time.Now().Format("20060102150405")
	className := "csi-hostpath-snapclass" // TODO: make it dynamic

	c.exsnapclient.SnapshotV1()

	snapTemplate := &snapshot.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", pvc.Name, currentTime),
			Namespace: pvc.Namespace,
		},
		Spec: snapshot.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &className,
			Source: snapshot.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvc.Name,
			},
		},
	}

	// fmt.Println("Snap template: ", *snapTemplate)

	_, err := c.exsnapclient.SnapshotV1().VolumeSnapshots(pvc.Namespace).Create(context.Background(),
		snapTemplate, metav1.CreateOptions{})

	if err == nil {
		log.Printf("Info: Volume Snapshot taken: %s-%s\n", pvc.Name, currentTime)
	}

	return err
}

// Handle CRD created event and put CRD obj in queue
func (c *Controller) HandleAddEvent(obj interface{}) {
	log.Println("Info: In HandleAddEvent")
	c.queue.Add(obj)
}
