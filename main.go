package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	// v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/streadway/amqp"
)

func GetClientSet() *kubernetes.Clientset {

	inCluster, exists := os.LookupEnv("RAS_IN_CLUSTER")
	if !exists {
		inCluster = "false"
		log.Printf("Running in cluster? %v", inCluster)
	}

	var clientset *kubernetes.Clientset
	var kubeConfig *rest.Config
	var err error

	if inCluster == "true" {
		log.Println("Creating an in cluster go-client config")
		kubeConfig, err = rest.InClusterConfig()

		if err != nil {
			// panic(err.Error())
			log.Printf("NOT GOOD: %v\n", err.Error())
		}

	} else {
		log.Println("Creating go-client config from user homedir")
		userHomeDir, err := os.UserHomeDir()
		must(err)

		kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		must(err)
	}

	clientset, err = kubernetes.NewForConfig(kubeConfig)
	must(err)

	return clientset
}

func must(err error) {
	if err != nil {
		log.Printf("This is not looking good! %v", err.Error())
		log.Panicf("Oops err: %v ", err)
	}
}

func main() {
	namespace, _ := os.LookupEnv("RAS_NAMESPACE")
	connectionString, _ := os.LookupEnv("RAS_CONNECTION_STRING")
	queueName, _ := os.LookupEnv("RAS_QUEUE")
	consumerDeploymentName, _ := os.LookupEnv("RAS_CONSUMER_DEPLOYMENT")

	k8sClientSet := GetClientSet()

	// Connect to RabbitMQ
	conn, err := amqp.Dial(connectionString)
	must(err)
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	must(err)
	defer ch.Close()

	// Declare the queue
	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	must(err)

	// Consumer Deployment
	consumerDeployment, err := k8sClientSet.AppsV1().Deployments(namespace).Get(context.TODO(), consumerDeploymentName, metav1.GetOptions{})
	must(err)

	//
	// The Loop
	//
	for {

		queueInfo, err := ch.QueueInspect(q.Name)
		must(err)

		fmt.Printf("queue: %v \tlength:  %v\n", q.Name, queueInfo.Messages)
		fmt.Printf("deployment %v\n", consumerDeployment.Name)

		var replicas int32
		switch {
		case queueInfo.Messages > 100:
			fmt.Println("Scale to 10")
			replicas = 10
		case queueInfo.Messages > 50:
			fmt.Println("Scale to 5")
			replicas = 5
		case queueInfo.Messages > 0:
			fmt.Println("Scale to 1")
			replicas = 1
		default:
			fmt.Println("Scale to 0")
			replicas = 0
		}

		// Update the deployment
		consumerDeployment.Spec.Replicas = &replicas
		_, err = k8sClientSet.AppsV1().Deployments(namespace).Update(context.TODO(), consumerDeployment, metav1.UpdateOptions{})

		time.Sleep(30 * time.Second)
	}
}
