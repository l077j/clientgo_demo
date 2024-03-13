package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	NAMESPACE       = "test-clientset"
	DEPLOYMENT_NAME = "client-test-deployment"
	SERVICE_NAME    = "client-test-service"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(option) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	operate := flag.String("operate", "create", "operate type: create or clean")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

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

func clean(clientset *kubernetes.Clientset) {
	emptyDeleteOptions := metav1.DeleteOptions{}

	if err := clientset.CoreV1().Services(NAMESPACE).Delete(context.TODO(), SERVICE_NAME, emptyDeleteOptions); err != nil {
		panic(err.Error())
	}

	if err := clientset.AppsV1().Deployments(NAMESPACE).Delete(context.TODO(), DEPLOYMENT_NAME, emptyDeleteOptions); err != nil {
		panic(err.Error())
	}

	if err := clientset.CoreV1().Namespaces().Delete(context.TODO(), NAMESPACE, emptyDeleteOptions); err != nil {
		panic(err.Error())
	}
}

func createNamespace(clientset *kubernetes.Clientset) {
	namespaceClient := clientset.CoreV1().Namespaces()

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: NAMESPACE,
		},
	}

	result, err := namespaceClient.Create(context.TODO(), namespace, metav1.CreateOptions{})

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Create namespace %s \n", result.GetName())
}

func createService(clientset *kubernetes.Clientset) {
	serviceClient := clientset.CoreV1().Services(NAMESPACE)

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
	deploymentClient := clientset.AppsV1().Deployments(NAMESPACE)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: DEPLOYMENT_NAME,
		},
		Spec: appsv1.DeploymentSpec{
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

	result, err := deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Create deployment %s \n", result.GetName())
}
