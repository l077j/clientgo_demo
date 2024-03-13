package main

import (
	"context"
	"flag"

	/*
	   当我们在运行 GO 程序时，可以在命令行界面中通过添加参数来传递一些参数信息，这些参数信息将被程序用来控制业务行为或操作行为。
	   而 flag 包就是 Go 标准库中用于解析这些命令行参数的包。
	   在程序中，通常会在程序启动的时候使用 flag 包来读取和解析命令行参数，获取参数信息，并根据参数信息来控制程序的行为。
	   要理解 flag.Parse() 的作用，需要先了解 flag 包中的几个重要函数：
	   flag.String()：用于解析带有 string 类型参数的命令行参数，并将其存储到一个指定的 string 变量中。
	   flag.Int()：用于解析带有 int 类型参数的命令行参数，并将其存储到一个指定的 int 变量中。
	   flag.Bool()：用于解析带有 bool 类型参数的命令行参数，并将其存储到一个指定的 bool 变量中。
	   flag.Parse()：用于解析命令行参数，并更新变量的值。
	   以上函数都有类似的使用方法。它们的第一个参数是要解析的命令行参数的名称，第二个参数是该参数的默认值，第三个参数是参数的说明信息。
	*/
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/pointer"
)

// 这个是go-lint进行代码静态检查时对NAMESPACE常量的检查结果。
// 在Go语言中，如果在代码中声明的常量、变量、函数或方法使用了大写字母开头的名称，
// 则表示这是一个公开的（或导出的）对象，可以被其他的包引用。在这种情况下，Go的文档注释是非常重要的，
// 建议对导出的变量和函数都加上注释。
// 根据 Kubernetes 命名规范，命名空间的名称必须符合 RFC 1123 标准，即名称由小写字母、数字和 "-"（减号）组成，长度不能超过 253 个字符。此外，名称不能以 "-" 开头或结尾。
// 因此，当你尝试创建名称为 "test_clientset" 的命名空间时，系统会报告这个错误，因为命名空间的名称以 "_clientset" 结尾，不符合命名空间名称的规则。
const (
	NAMESPACE       = "test-clientset"
	DEPLOYMENT_NAME = "client-test-deployment"
	SERVICE_NAME    = "client-test-service"
)

func main() {
	var kubeconfig *string
	// home是家目录，如果能够取得家目录的值，就可以用来做默认值
	if home := homedir.HomeDir(); home != "" {
		// 如果输入了kubeconfig参数，该参数的值就是kubeconfig文件的绝对路径
		// 如果没有输入kubeconfig参数，使用默认路径~/.kube/config
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		// 如果取不到当前目录的家目录，就没办法设置kubeconfig的默认目录了，只能从入参中取
		kubeconfig = flag.String("kubeconfig", "", "absolute paht to the kubeconfig file")
	}
	/*如果可以获取到用户家目录 home，则使用 filepath.Join() 方法获得默认的 kubeconfig 路径，
	  即 ${HOME}/.kube/config，并将值赋给 kubeconfig 参数。
	  如果无法获取到用户家目录 home，则将 kubeconfig 的默认值设置为空字符串。
	  这时候程序会尝试从环境变量 KUBECONFIG 中读取默认的 kubeconfig 路径，如果无法找到，程序将使用 Kubernetes 集群的默认路径 ${HOME}/.kube/config
	*/

	// 获取用户输入的操作类型，默认是create，可以手动输入，如clean，用于清理所有资源
	// operate 是一个存储操作类型值的变量。程序通过读取用户在命令行中输入的操作类型的值（也就是输入的参数值），
	// 并把它保存到 operate 这个变量中，来获取从命令行输入的操作类型。
	// 在 Go 语言中，flag.String() 用于解析一个字符串类型的命令行参数，并返回一个指针，该指针指向解析参数所存储值的引用
	operate := flag.String("operate", "create", "operate type : create or clean")

	flag.Parse()
	// flag.Parse() 函数来解析命令行参数，这个函数会遍历 os.Args 切片，并根据类型解析每个参数值。在解析每个参数值后，
	// flag.Parse() 会将解析结果存储到对应的变量中，使我们能够在程序中使用这些变量，读取和控制命令行参数对程序的影响

	/*
	   clientcmd 是 Kubernetes Go 客户端的一部分，是 client-go/tools/clientcmd 包中的一个子包，
	   用于加载和生成 kubernetes/client-go/rest.Config 配置对象。该包提供了一些函数，
	   用于从不同的来源获取 kubeconfig 文件、加载配置并生成 Kubernetes Rest 客户端所需的 Config 配置对象。
	   BuildConfigFromFlags() 是 clientcmd 包中的一个函数，用于从命令行标志或环境变量中解析 kubeconfig 文件的路径，
	   然后调用 BuildConfigFromKubeconfigPath() 或 NewNonInteractiveDeferredLoadingClientConfig() 函数生成 rest.Config 类型的对象。
	   在这个代码中，clientcmd.BuildConfigFromFlags() 函数的作用是从命令行标志和环境变量中解析获取 kubeconfig 文件路径，
	   并生成一个 rest.Config 类型的对象。这个函数的第一个参数是指定的 kubeconfig 文件路径，这里为空字符串，表示使用默认的 kubeconfig 文件路径。
	   第二个参数为指向 kubeconfig 文件路径的指针，它是一个命令行标志参数的值，指向命令行标志变量 kubeconfig 的指针，
	   表示从命令行参数或环境变量获取 kubeconfig 文件路径。通过调用 clientcmd.BuildConfigFromFlags() 函数，程序会获取可用的 kubeconfig 文件路径，
	   并将它作为参数来创建一个 Kubernetes 客户端配置对象 config。这个配置对象包含了与 Kubernetes API Server 连接所需的所有数据，
	   包括 API Server 地址、身份验证信息等，用于与 Kubernetes API Server 进行通信。
	*/
	// 从本机加载kubeconfig配置文件，因此第一个参数为空字符串
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		panic(err.Error())
	}

	/*
	   kubernetes 是 client-go/kubernetes 库的一部分，它提供了访问 Kubernetes API 的客户端接口。
	   这个库中的类型和函数定义了用于操作 Kubernetes 集群的客户端实现。kubernetes 这个子包中有很多 Kubernetes API 可以进行访问，例如 Pods、Services、Deployments 等。
	   kubernetes.NewForConfig() 是这个客户端库中的一个函数，用于为指定的 Kubernetes 配置信息（即 config）创建一个 Kubernetes 客户端集合（即 clientset）对象。
	   这个客户端集合对象包含了可以访问 Kubernetes API 的函数和方法。这个函数会返回一个新创建的 kubernetes.Interface 类型的对象，其中包含了用于与 Kubernetes API Server 进行连接的客户端、方法和配置。
	   在这个代码中，程序使用 kubernetes.NewForConfig() 函数，从之前通过 clientcmd.BuildConfigFromFlags() 函数返回的 Kubernetes 客户端配置对象 config 中获取 Kubernetes 配置信息，并根据这些配置信息创建 Kubernetes 客户端集合对象 clientset。
	   这个客户端集合对象包含了访问 Kubernetes API 的所有功能，并且可以向 Kubernetes API Server 发送请求以对 Kubernetes 资源进行管理和处理。
	   这个 clientset 对象使得程序可以使用 Go 语言来管理 Kubernetes 集群，例如获取集群中的各种资源对象的状态和信息，创建、更新和删除这些资源对象等。这些操作都可以通过调用 clientset 中的各种方法和函数，来访问 Kubernetes API Server 和管理 Kubernetes 资源。
	*/
	// 实例化clientset对象
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("operate is %v\n", *operate)

	if "clean" == *operate {
		clean(clientset)
	} else {
		createNamespace(clientset)
		createDeployment(clientset)
		createService(clientset)
	}

}

/*
具体来说，该函数（clean）使用传递进来的 Kubernetes 客户端集合 clientset，调用 Kubernetes API Server 中与 Job/Pod 操作相关的函数和方法。
在函数内部，它调用了 clientset.BatchV1().Jobs(namespace).List() 方法，以获取 namespace 命名空间中所有的 Job 集合，
然后使用 clientset.BatchV1().Jobs(namespace).Delete() 方法删除所有已完成的 Job。
删除作业时，它还会使用 clientset.CoreV1().Pods(namespace).Delete() 方法，删除与已完成作业相关的 Pod。
clean() 函数的作用是清理指定命名空间中已完成的任务和它们关联的 Pod，以释放 Kubernetes 集群资源并提高集群效率。
 这个函数通过传递 Kubernetes 客户端集合来完成，因此在编写这个函数时，需要先通过 kubernetes.NewForConfig() 函数创建 Kubernetes 客户端集合对象，并将其传递到这个函数中。
*/
// 参数是一个指向 kubernetes.Clientset 类型的指针 clientset，用于传递 Kubernetes 客户端集合。
// 这个函数的作用是删除指定命名空间中已经完成的 Job 和它们创建的 Pod。
func clean(clientset *kubernetes.Clientset) {

	/*这行代码的含义是创建了一个名为 emptyDeleteOptions 的变量，类型为 metav1.DeleteOptions{}，并将其初始化，使其为空。
	在 Kubernetes 中，当我们删除某个资源对象时，需要传递一组可选参数，用于指定如何删除这个对象。
	这些参数被定义在 metav1.DeleteOptions{} 类型中，并包括了以下字段：GracePeriodSeconds、Preconditions、PropagationPolicy 和 DryRun。
	如果删除资源对象时不需要使用任何可选参数，可以创建一个空的 metav1.DeleteOptions 对象，
	将其作为删除函数的参数传递给 Kubernetes 客户端函数，这样就可以使用默认设置删除资源对象*/
	// 在后续的代码中，这个空 DeleteOptions 对象被用于删除 Job 和 Pod 对象，因为这些资源对象都可以使用默认的参数来进行删除。
	// 另外，在不需要删除资源对象时，也可以直接使用 nil 值来指定删除参数，以达到与空的 DeleteOptions 对象相同的作用
	emptyDeleteOptions := metav1.DeleteOptions{}

	// 删除service
	/*
		clientset.CoreV1().Services(NAMESPACE).Delete(context.TODO(), SEVRICE_NAME, emptyDeleteOptions)：删除指定名称空间（NAMESPACE）中名为 SEVRICE_NAME 的 Service 资源对象。
		该操作使用 CoreV1() 方法来获取核心 API 的资源对象，并使用 Services(NAMESPACE) 方法来访问该命名空间下的 Service 列表。
		删除 Service 对象时，使用 Delete() 方法，传递 emptyDeleteOptions 作为删除选项，以使用默认设置从 Kubernetes API Server 删除资源对象。如果删除失败，则会出现一个 panic，将错误信息输出到控制台
	*/
	if err := clientset.CoreV1().Services(NAMESPACE).Delete(context.TODO(), SERVICE_NAME, emptyDeleteOptions); err != nil {
		panic(err.Error())
	}

	// 删除deployment
	/*
		clientset.AppsV1().Deployments(NAMESPACE).Delete(context.TODO(), DEPLOYMENT_NAME, emptyDeleteOptions)：删除指定名称空间（NAMESPACE）中名为 DEPLOYMENT_NAME 的 Deployment 资源对象。
		该操作使用 AppsV1() 方法来获取应用程序 API 的资源对象，并使用 Deployments(NAMESPACE) 方法来访问该命名空间下的 Deployment 列表。
		删除 Deployment 对象时，使用 Delete() 方法并传递 emptyDeleteOptions。如果删除失败，则会出现一个 panic，将错误信息输出到控制台
	*/
	if err := clientset.AppsV1().Deployments(NAMESPACE).Delete(context.TODO(), DEPLOYMENT_NAME, emptyDeleteOptions); err != nil {
		panic(err.Error())
	}

	// 删除namespace
	/*
		clientset.CoreV1().Namespaces().Delete(context.TODO(), NAMESPACE, emptyDeleteOptions)：删除名为 NAMESPACE 的命名空间。
		该操作使用 CoreV1() 方法来获取核心 API 的资源对象，并使用 Namespaces() 方法来访问 Kubernetes 中所有的命名空间资源对象。
		删除 Namespace 对象时，使用 Delete() 方法并传递 emptyDeleteOptions。如果删除失败，则会出现一个 panic，将错误信息输出到控制台。
	*/
	if err := clientset.CoreV1().Namespaces().Delete(context.TODO(), NAMESPACE, emptyDeleteOptions); err != nil {
		panic(err.Error())
	}
}

func createNamespace(clientset *kubernetes.Clientset) {
	// 通过调用 clientset.CoreV1().Namespaces() 来获取命名空间客户端的函数，我们可以获得一个用于创建和操作命名空间资源的客户端，并通过调用它访问 Kubernetes API 中的命名空间，以实现对命名空间资源的操作。
	/*
		CoreV1() 方法用于访问 Kubernetes 核心 API 的资源对象。
		Namespaces() 方法用于获取命名空间资源对象并创建与之交互的 Kubernetes 客户端。
		clientset 是之前通过 kubernetes.NewForConfig() 函数创建的 Kubernetes 客户端集合对象，它提供了与 API Server 通信的便捷方法和函数。
		namespaceClient 是一个 NamespaceInterface 类型的对象，它用于操作 Kubernetes 中的命名空间资源，例如创建、更新和删除命名空间资源等。
	*/
	namespaceClient := clientset.CoreV1().Namespaces()

	// 定义了一个变量 namespace，它是一个指向 apiv1.Namespace 类型对象的指针。这个变量用于存放新创建的命名空间的元数据信息。
	/*
		ObjectMeta 包含 Kubernetes API 资源对象的元数据信息，这里指定了要创建的命名空间的名称。
		Name 表示要创建的命名空间的名称，它是一个常量 NAMESPACE。
		metav1.ObjectMeta 是一个具有元数据的对象，用于定义 Kubernetes 资源对象的基本信息。在这里，我们指定这个 ObjectMeta 对象的属性为新命名空间的名称。
	*/
	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: NAMESPACE,
		},
	}

	// 通过调用 namespaceClient.Create() 方法来创建一个新的命名空间，并将其存储在 namespace 变量中
	// 使用客户端集合和命名空间客户端 namespaceClient 来创建一个新的 Kubernetes 命名空间对象，并返回一个包含命名空间详细信息的 corev1.Namespace 对象（result）以及任何可能发生的错误（err）
	/*
		context.TODO() 是一个空的上下文，代表函数不需要任何特殊的上下文信息。
		namespace 是我们之前定义的用于存放新命名空间元数据信息的 Namespace 对象，它指定了新命名空间的名称和其他元数据。
		metav1.CreateOptions{} 表示在创建命名空间时不传递任何额外的选项和参数。
	*/
	result, err := namespaceClient.Create(context.TODO(), namespace, metav1.CreateOptions{})

	if err != nil {
		panic(err.Error())
	}

	// %s 表示字符串参数
	fmt.Printf("Create namespace %s \n", result.GetName())
}

func createService(clientset *kubernetes.Clientset) {
	// 使用 CoreV1() 函数获取 Kubernetes API 中 Core API 资源对象的客户端集合，然后使用 Services(NAMESPACE) 方法访问 NAMESPACE 命名空间中的所有服务（Services）资源对象，并创建与之交互的 Kubernetes 客户端。
	/*
		clientset 是之前通过 kubernetes.NewForConfig() 函数创建的 Kubernetes 客户端集合对象，它提供了与 API Server 通信的便捷方法和函数。
		通过使用 clientset.CoreV1() 方法来获取 Kubernetes Core API 相关的客户端方法和函数。
		serviceClient := clientset.CoreV1().Services(NAMESPACE) 中的 Services(NAMESPACE) 方法用于获取 Kubernetes API Server 中与服务资源对象相关的客户端集合，并且指定命名空间（NAMESPACE）用于限制服务的范围。
	*/
	serviceClient := clientset.CoreV1().Services(NAMESPACE)

	// 定义了一个 apiv1.Service 类型的指针变量 service，用于存储已定义的新服务资源对象的元数据信息。
	/*
		ObjectMeta 包含 Kubernetes API 资源对象的元数据信息，这里指定了要创建的服务的名称为 SERVICE_NAME。
		Name 表示要创建的服务的名称，它是一个常量 SERVICE_NAME。
		Spec 表示服务的详细信息，包括所使用的端口、服务类型（NodePort）和选择器，以及与服务可能关联的其它信息。
		Ports 表示服务监听的端口号，它只有唯一一项，是用于监听 HTTP 流量的，命名为 "http"，容器端口号为 8080，然后将它们绑定到节点的 30080 端口。
		Selector 指定了将要选择的标签，以便建立与端点 Pod 的关联，这里定义了一个标签 (app:tomcat)。
		Type 表示 Kubernetes 服务类型，这里使用的是 apiv1.ServiceTypeNodePort，可以使用任何类型的 Kubernetes 服务类型，根据特定的应用程序需要选择不同类型的服务对象。
	*/
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: SERVICE_NAME,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{{
				Name:     "http",
				Port:     8080,
				NodePort: 30080,
			},
			},
			Selector: map[string]string{
				"app": "tomcat",
			},
			Type: apiv1.ServiceTypeNodePort,
		},
	}

	result, err := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Create service %s \n", result.GetName())

}

func createDeployment(clientset *kubernetes.Clientset) {
	// 使用 Kubernetes 客户端对象集合和应用程序 API 资源对象的客户端方法和函数，来访问和管理 Kubernetes 部署资源对象。
	/*
		clientset 是之前通过 kubernetes.NewForConfig() 函数创建的 Kubernetes 客户端集合对象，它提供了与 API Server 通信的便捷方法和函数。
		通过调用 AppsV1() 方法来获取 Kubernetes 应用程序 API 的客户端方法和函数。
		deploymentClient := clientset.AppsV1().Deployments(NAMESPACE) 中的 Deployments(NAMESPACE) 方法用于获取 Kubernetes API Server 中与部署资源对象相关的客户端集合，并且指定命名空间（NAMESPACE）用于限制部署操作的范围。
	*/
	deploymentClient := clientset.AppsV1().Deployments(NAMESPACE)

	// 定义了一个指向 appsv1.Deployment 类型的指针变量 deployment，用于存储要创建的新部署资源的元数据信息。
	/*
		ObjectMeta 包含 Kubernetes API 资源对象的元数据信息，这里指定了要创建的部署资源对象的名称为 DEPLOYMENT_NAME。
		Name 表示要创建的部署资源对象的名称，它是一个常量 DEPLOYMENT_NAME。
		Spec 表示部署的详细信息，包括复制数、选择器、定义 Pod 模板等等。
		Replicas 表示需要部署的 Pod 的个数，这里设置为 2。
		Selector 指定了部署的机制，以此用于根据特定的选择器匹配进入部署的每一个 Pod 实例。在这里选择了选择器 app:tomcat。
		Template 是一个定义在部署之中的 Pod 模板，提供容器的元数据和其他相关信息，可以用恰当的方式定义出必要的容器属性等。
		Labels 定义了模板中容器的元数据，用于匹配 Selector 中的标签，以便向部署中添加 Pod。在这里选择了选择器 app:tomcat。
		Containers 是包含部署中容器的列表，每个容器都有一个预定义的设置，如容器名称、镜像名称、端口号等
	*/
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: DEPLOYMENT_NAME,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "tomcat",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "tomcat",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "tomcat",
							Image:           "tomcat:8.0.18-jre8",
							ImagePullPolicy: "IfNotPresent",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolSCTP,
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
		},
	}

	// 调用 deploymentClient.Create() 方法来将定义的新部署资源对象 deployment 存储在 Kubernetes 中，并将相关参数传递给此函数。
	/*
		context.TODO() 是一个空的上下文，代表函数不需要任何特殊的上下文信息。虽然，在某些情况下，调用函数时必须传递一个上下文参数，例如取消请求和处理超时等。但是，在本例里，我们使用了一个简单的 TODO() 空上下文。
		deployment 是用于存储新部署资源对象元数据信息的 Deployment 对象。这个对象包含了要部署的副本数、所使用的选择器和相关其他信息，用于标识和管理该部署。
		这个参数是通过之前定义的指向 appsv1.Deployment 类型的指针来传递的，以便在 Kubernetes 中使用。
		metav1.CreateOptions{} 表示创建部署资源对象时不需要传递任何附加的选项和参数。这个参数用于传递 Kubernetes 资源对象的附加选项信息和其他注释信息等上下文参数。
	*/
	result, err := deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Create deployment %s \n", result.GetName())
}
