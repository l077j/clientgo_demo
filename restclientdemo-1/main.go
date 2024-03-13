package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

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

	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err.Error())
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	result := &corev1.PodList{}

	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("namespace\t status\t name\n")
	for _, ns := range namespaces.Items {
		namespace := ns.Name

		err = restClient.Get().Namespace(namespace).Resource("pods").VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).Do(context.TODO()).Into(result)
		if err != nil {
			panic(err.Error())
		}

		for _, d := range result.Items {
			fmt.Printf("%v\t %v\t %v\n", d.Namespace, d.Status.Phase, d.Name)
		}
	}

}
