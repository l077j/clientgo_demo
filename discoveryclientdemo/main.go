package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// DiscoveryClient 客户端是发现客户端，主要用于发现 Kubernetes API Server 所支持的资源组、资源版本、资源信息。
func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(option) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		panic(err.Error())
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	APIGroup, APIResourceListSlice, err := discoveryClient.ServerGroupsAndResources()

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("APIGroup:\n\n %v\n\n\n\n", APIGroup)

	for _, singleAPIResourceList := range APIResourceListSlice {
		groupVersionStr := singleAPIResourceList.GroupVersion
		gv, err := schema.ParseGroupVersion(groupVersionStr)

		if err != nil {
			panic(err.Error())
		}

		fmt.Println("************************************************************")
		fmt.Printf("GV string [%v]\nGV struct [%#v]\nresources:\n\n", groupVersionStr, gv)

		for _, singleAPIResource := range singleAPIResourceList.APIResources {
			fmt.Printf("%v\n", singleAPIResource.Name)
		}
	}
}
