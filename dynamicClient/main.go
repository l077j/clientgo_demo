package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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

	// 通过 k8s.io/client-go/dynamic 包中的 NewForConfig() 方法创建 DynamicClient 对象，该对象可以与 Kubernetes API 中的动态 API 资源交互。
	dynamicClient, err := dynamic.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	// 定义一个 GroupVersionResource 类型变量 gvr，用于指示 Kubernetes API 中的资源对象、组和版本。
	/*
		schema 是 Kubernetes API 的积极协调协议（GroupVersionKind）组件，用于在线路上传输和管理 Kubernetes 资源对象的版本、组、类型和其他元数据。
		GroupVersionResource 包含用于唯一标识 Kubernetes API 中资源对象的组、版本和资源路径信息。在这里，它们代表了 Kubernetes API 中的 Group-Resource-Version 元数据。
		Version 告诉 API Server 您要检索的资源的 Kubernetes API 版本号（本例中是 v1）。
		Resource 告诉 API Server 您要检索的资源的 Kubernetes API 组（group name），并且限定了访问的资源（例如，在本例中为 "pods"，代表 Pod 资源对象）。
	*/
	gvr := schema.GroupVersionResource{Version: "v1", Resource: "pods"}

	// 使用之前创建的 DynamicClient 对象来获取 Kubernetes 集群中指定版本组和资源的动态资源对象的列表。
	/*
		dynamicClient.Resource(gvr) 用于检索指定 GroupVersionResource 的资源对象列表。
		在这里，我们传递了由 schema.GroupVersionResource{Version: "v1", Resource: "pods"} 定义的变量 gvr，这表示我们想要访问 Kubernetes API 中的 Pod 资源对象列表。
		Namespace() 方法是可选的，它用于设置 Kubernetes 命名空间。在这个例子中，我们为命名空间设置了 "kube-system"，因此，只会返回该命名空间的 Pod。
		List() 方法按照给定的列表选项列出 Kubernetes Pod 资源对象列表。此方法通过参数 context.TODO() 展示，没有特殊的上下文要求。在这里，我们使用 Limit 选项来限制列表返回的结果数量。
	*/
	unstructObj, err := dynamicClient.
		Resource(gvr). // Resource 方法指定了本次操作的资源类型
		Namespace("kube-system").
		List(context.TODO(), // List 方法向kubernetes发起请求
			metav1.ListOptions{Limit: 100})

	if err != nil {
		panic(err.Error())
	}

	podList := &apiv1.PodList{}

	// FromUnstructured 将 Unstructured 数据结构转成 PodList
	// 使用 Kubernetes 的默认结构转换器 DefaultUnstructuredConverter，将 unstructured.UnstructuredList 类型的对象 unstructObj 转换为 PodList 类型的对象 podList
	/*
		FromUnstructured() 方法将类型为 unstructured.UnstructuredList 的 unstructObj 转换为 Kubernetes API 中 PodList  对象类型 podList。
		这种类型转换主要使用 kubeapi 包的 runtime 包中的 interface ToUnstructured(obj interface{}) 和 interface FromUnstructured(data map[string]interface{}, obj interface{}) 实现。
		由于 unstructObj 中的数据是动态资源的原始版本，需要使用转换器将其转换为已知的 PodList 类型。
		转换器将特定数据结构转换为特定类型的数据结构，从而使我们可以将繁琐和复杂的过程抽象出来，转换为可读和易于管理的类型和结构，并简化我们的代码。
		FromUnstructured() 方法将获取到的 unstructured.UnstructuredList 格式数据转换为强类型的 PodList 对象，该对象具有 Kubernetes API 中 PodList 对象的结构和元数据形式。
	*/
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructObj.UnstructuredContent(), podList)

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("namespace\t status\t\t name\n")

	for _, d := range podList.Items {
		fmt.Printf("%v\t %v\t\t %v\n",
			d.Namespace,
			d.Status.Phase,
			d.Name)
	}
}
