package main

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"strings"
)

// Make a hash map that stores known pods.
var knownPods map[string]string

func getAllPods() *v1.PodList {
	// use the current context in the kubeconfig file
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{})

	// get a config object
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		log.Printf("Could not get K8s client: %v", err)
		return nil
	}

	// create a clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Could not create K8s config: %v", err)
		return nil
	}

	// get all pods in all namespaces
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Could not get K8s namespaces: %v", err)
		return nil
	}

	return pods
}

// getPodFromContainer retrieves pod from container's ID. This shall be optimized in the future..
func getPodFromContainer(containerID string) string {
	// Get the container id ignoring docker prefix
	containerID = strings.Replace(containerID, "docker-", "", 1)

	// If the container's id is already registered in the map, return pod name.
	if _, ok := knownPods[containerID]; ok {
		return knownPods[containerID]
	} else { // If the container's id does not exist in the map, do dirty job O(n^2)
		// Get all pods from K8s and retrieve the pods and container information
		pods := getAllPods()
		if pods == nil {
			log.Printf("Could not get information from K8s")
		}

		// Do the dirty job.
		for _, pod := range pods.Items {
			for _, stat := range pod.Status.ContainerStatuses {
				if strings.Contains(stat.ContainerID, containerID) {
					knownPods[containerID] = pod.Name
					return pod.Name
				}
			}
		}
	}

	return ""
}
