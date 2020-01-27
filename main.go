package main

import (
	"log"
	"os"

	batchV1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	clientset *kubernetes.Clientset
)

func init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("ERROR: init(): Could not get kube config in cluster. Error:" + err.Error())
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("ERROR: init(): Could not connect to kube cluster with config. Error:" + err.Error())
	}
}

func main() {
	cleanJobs(os.Getenv("NAMESPACE"))
	cleanPods(os.Getenv("NAMESPACE"))
}

// cleanJobs - Filters for successful completions
// Trigger brute force filter for failures, with risk of out of memory. By removing completed jobs first, this risk is reduced
// Risk of out of memory is also managed by the whole reason of this program: Hitting a 10k job limit
func cleanJobs(namespace string) {
	removeJobs(namespace, "status.successful=1")
	removeJobs(namespace, "")
}

func removeJobs(namespace string, filter string) {
	list, err := clientset.BatchV1().Jobs(namespace).List(metav1.ListOptions{FieldSelector: filter})
	if err != nil {
		log.Fatalf("ERROR: cleanJobs(): Can not list jobs for namespace %s. Error %v", namespace, err)
	}
	// Clean up the jobs
	for _, x := range list.Items {
		// inspect job status: No pods should be active (status.active=0), job should be complete or failed
		if x.Status.Active == 0 {
			for _, s := range x.Status.Conditions {
				if s.Type == batchV1.JobComplete || s.Type == batchV1.JobFailed {
					log.Printf("INFO: cleanJobs(): Deleting job %s with status %s", x.Name, s.Type)
					if err := clientset.BatchV1().Jobs(namespace).Delete(x.Name, &metav1.DeleteOptions{}); err != nil {
						log.Printf("INFO: cleanJobs(): failed to delete job %s. Error %v", x.Name, err)
					}
				}
			}
		}
	}
}

// cleanPods - Wraps removePods to add filters reducing memory use and passing filtering load to kubernetes master
func cleanPods(namespace string) {
	removePods(namespace, "status.phase=Succeeded")
	removePods(namespace, "status.phase=Failed")
}

// removePods - Removes pods from kubernetes cluster according to provided filter
func removePods(namespace string, filter string) {
	list, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{FieldSelector: filter})
	if err != nil {
		log.Fatalf("ERROR: cleanPods(): Can not list pods for namespace %s. Error %v", namespace, err)
	}
	// Clean up the pods
	for _, x := range list.Items {
		// inspect pod status:
		log.Printf("INFO: cleanPods(): pod %s with status %s removed", x.Name, x.Status.Phase)
		if err := clientset.CoreV1().Pods(namespace).Delete(x.Name, &metav1.DeleteOptions{}); err != nil {
			log.Printf("INFO: cleanPods(): failed to delete pod %s. Error %v", x.Name, err)
		}
	}
}
