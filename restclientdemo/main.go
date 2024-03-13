package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	/*
		import corev1 "k8s.io/api/core/v1" 的作用是导入 k8s.io/api/core/v1 这个包，并为其设置一个短路径 corev1，以方便在代码中使用。
		换句话说，这个语句的作用是将 k8s.io/api/core/v1 这个绝对路径的长包名，映射为相对于这个代码包的短包名 corev1。
		这样，在代码中，可以通过使用 corev1 作为前缀，来访问和使用 k8s.io/api/core/v1 包里面定义的所有类型和函数。
	*/
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
	// 用 homedir 库获取当前用户的家目录路径,HomeDir()函数会查找环境变量
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		// flag.String() 是 Go 语言标准库 flag 包中用来解析命令行参数的函数之一，它的作用是从命令行读取字符串类型的参数值
		// filepath.Join() 函数会把它的参数用操作系统特定的路径分隔符拼接成一个完整的路径
		// filepath.Join(home, ".kube", "config") 这个路径就是 kubeconfig 文件的默认路径，即 ${HOME}/.kube/config
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse() // flag 包中的函数，用于解析命令行参数。命令行参数是指在终端输入参数
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// 使用 clientcmd 库来加载 kubeconfig 文件，并返回一个 Kubernetes 客户端配置对象 config
	// 是从指定 kubeconfig 文件路径参数中加载 Kubernetes API Server 的连接参数，并生成一个Kubernetes客户端的配置对象 config
	// BuildConfigFromFlags 函数使用 client-go 库来读取和解析 kubeconfig 文件

	if err != nil {
		panic(err.Error())
		// panic() 函数是用于抛出运行时异常的 Go 语言内置函数。当程序执行到 panic() 时，程序会立即停止运行，然后将错误信息打印到控制台，并输出调用栈信息，以便更好地追踪和调试问题
		// err.Error() 表示将错误信息转化为字符串类型输出
	}

	// 初始化 Kubernetes REST Client 的配置信息
	config.APIPath = "api" //设置 REST API 的路径信息
	config.GroupVersion = &corev1.SchemeGroupVersion
	// 设置REST API的Group和Version信息,在这个代码中,选择的Group为corev1,Pod对象的Group是空字符串,Version为v1
	config.NegotiatedSerializer = scheme.Codecs
	// 指定序列化方法。这个变量保存的是client-go/kubernetes/scheme 包中定义的一个序列化编解码器的实例，
	// 用于将HTTP POST和PUT操作的body请求体序列化/反序列化成 Go 语言中的结构体类型
	// 在这个代码中，使用 scheme.Codecs 表示使用 Kubernetes 标准定义的序列化方式

	restClient, err := rest.RESTClientFor(config)
	// 使用 config 对象里面设置的与 Kubernetes API Server 连接的参数，构建Kubernetes的REST Client客户端对象 restclient
	// 调用了 Kubernetes 的 client-go/rest 库提供的 RESTClientFor(config) 方法
	// 这个 REST Client 对象包括了发送 HTTP 请求的功能和必需的身份验证信息，是管理 Kubernetes 资源和状态的核心组件之一

	if err != nil {
		panic(err.Error())
	}

	result := &corev1.PodList{}
	// 定义一个类型为 corev1.Pod 的变量 result，并用 Go 语言的 & 符号创建一个指向这个变量地址的指针

	// var namespace string
	// namespace := "kube-system"
	clientSet, err := kubernetes.NewForConfig(config)

	/*
		clientset.CoreV1()：通过前面生成的 Kubernetes 客户端集合对象 clientset，获取与 Kubernetes 核心 API 相关的客户端方法。
		clientset.CoreV1().Namespaces()：通过链式调用，访问 Kubernetes 核心 API 中的 Namespaces 资源对象，以获得所有命名空间的信息和状态。
		.List(context.TODO(), v1.ListOptions{})：调用 API Server 中 Namespaces 对象的列表操作，获取所有的命名空间。ListOptions{} 参数为空，表示不需要设置任何特殊选项或筛选条件，获取所有的命名空间对象列表。
		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})：将返回的 corev1.NamespaceList 对象存储在 namespaces 变量中，并捕获任何发生的错误信息。
		这个代码片段中存储了 contexts.TODO() 上下文，它表示这个函数不需要特别的上下文信息。
	*/
	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	// 打印表头
	// 当使用%v格式化占位符时，Go会按照默认格式打印值，这可能是结构体、数组、切片、映射、接口等类型的值，而不需要显示地指定格式。Go会根据变量的类型来自动选择最适合的打印格式。
	fmt.Printf("namespace\t\t status\t\t name\n")

	for _, ns := range namespaces.Items {
		namespace := ns.Name
		err = restClient.Get().
			// 指定namespace
			Namespace(namespace).
			// 查找多个pod
			Resource("pods").
			// 指定大小限制和序列化工具，指定要使用的 URL 参数并指定其版本，设置返回的 Pod 个数限制为 100 个
			VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).
			// 请求
			Do(context.TODO()).
			// 结果存入result
			Into(result)

			/*
			   1 restClient.Get()
			   这个函数用于向 Kubernetes API Server 发送一个 GET 请求对象。
			   它返回一个包含 REST 调用对象和请求上下文信息的 rest.Request 类型变量。
			   2 .Namespace(namespace)
			   这是在请求对象上设置资源所属命名空间的方法。
			   Namespace(namespace) 设置要请求的 Kubernetes 资源的命名空间，比如 Pods、Services 和 Deployments。
			   3 .Resource("pods")
			   这个函数用于设置 REST 请求中资源的类型和版本信息。
			   Resource("pods") 表示要查询的资源类型为 pods 的 Kubernetes 资源。
			   4 .VersionedParams(&metav1.ListOptions{Limit: 100}, scheme.ParameterCodec)
			   这个函数用于设置 REST 请求的 URL 参数和版本信息。
			   这里使用了 VersionedParams 函数设置 URL 参数中的 Limit 最大数量为 100，表示最多只返回 100 个 Pod。
			   scheme.ParameterCodec 是 Kubernetes 内部使用的编解码器，用于将 Go 语言对象转换为 URL 参数形式。
			   5 .Do(context.TODO())
			   这个函数是发送 GET 请求，执行这个方法实际上返回的是一个 error type 值，用于检查错误是否发生。
			   如果没有发生错误，就可以继续从响应体中读取数据。参数 context.TODO() 表示使用默认的上下文信息，即不需要使用更加详细的上下文信息。
			   6 .Into(result)
			   这个函数将请求响应信息存储到变量 result 中。它使用 Go 语言的 & 操作符创建一个取址器（即指针）的方式，将值存储到变量 result 中。这个变量必须在上面首先被声明，并且指针类型必须与方法返回的类型相匹配。在代码中，& 表示创建 result 变量的地址，后面的值则表示存储响应数据的变量。
			   这些函数和方法在 client-go/rest 库中定义，用于简化向 Kubernetes API Server 发送 GET 请求时的代码实现。
			   在使用它们时，只需要按照需求设置参数，然后直接调用相应的函数即可。
			*/

		if err != nil {
			panic(err.Error())
		}

		// 每个pod都打印namespace,status.Phase,name三个字段
		for _, d := range result.Items {
			fmt.Printf("%v\t %v\t %v\n",
				d.Namespace,
				d.Status.Phase,
				d.Name)
		}
		// 循环器会在每次迭代时返回当前处理的 Pod 对象 d，_ 表示占位符
	}

}
