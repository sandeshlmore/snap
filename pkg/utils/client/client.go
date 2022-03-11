package client

import (
	"flag"
	"log"
	"path/filepath"

	snapclient "github.com/db/snap/pkg/generated/clientset/versioned"
	exsnapclientset "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// get clientsets
func GetKubeClient() (*kubernetes.Clientset, *snapclient.Clientset, *exsnapclientset.Clientset) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Println("Info: Error loading kube configuration from directory ", err)
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Println("Error: colud not load config from inclusterconfig ", err.Error())
		}
		log.Println("Info: Loading config from cluster sucessful. ")
	}
	// log.Println("config: ", config)

	// create kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("getting std client %s\n", err.Error())
	}

	// create the clientset for snapshot
	snapclientset, err := snapclient.NewForConfig(config)
	if err != nil {
		log.Printf("getting snapshot client %s\n", err.Error())
		panic(err.Error())
	}

	// create the clientset for external-snapshot
	exsnapclient, err := exsnapclientset.NewForConfig(config)
	if err != nil {
		log.Printf("getting external-snapshot client %s\n", err.Error())
		panic(err.Error())
	}

	return client, snapclientset, exsnapclient
}
