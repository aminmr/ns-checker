package main

import (
	"bytes"
	// "context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	// "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	
)

// Namespace represents a Kubernetes namespace
type Namespace struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}

// NamespaceList represents a list of Kubernetes namespaces
type NamespaceList struct {
	Items []Namespace `json:"items"`
}

func main() {

	// Load kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// Fallback to default kubeconfig location
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error retrieving home directory:", err)
			return
		}
		kubeconfig = homeDir + "/.kube/config"
	}
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	// Get the namespace has been choosed in kubeconfig
	namespace, _, err := configLoader.Namespace()
	if err != nil {
		fmt.Println("Error retrieving namespace:", err)
		return
	}

	// // Create Kubernetes client
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	// 	fmt.Println("Error getting kubeconfig:", err)
	// 	return
	// }
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	fmt.Println("Error creating Kubernetes client:", err)
	// 	return
	// }

	// Pass through the rest of the command to kubectl
	args := os.Args[1:]

	cmd := exec.Command("kubectl", args...)

	// // List namespaces
	// namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	fmt.Println("It's not OK:", err)
	// 	return
	// }

	// 	if namespaces != nil {
	// 	fmt.Println("It's OK")
	// } else {
	// 	fmt.Println("It's not OK: No namespaces found")
	// }

	createNsPresent := false
	getNsPresent := false

	// Check for 'create ns' or 'get ns' in all positions of args
	for i := 0; i < len(args)-1; i++ { // -1 to avoid index out of range on args[i+1]
		if args[i] == "create" && args[i+1] == "ns" {
			createNsPresent = true
			break // Exit the loop if found
		} else if args[i] == "get" && args[i+1] == "ns" {
			getNsPresent = true
			break // Exit the loop if found
		}
	}
	if createNsPresent || getNsPresent {
		// Execute the original kubectl command
		cmd.Stdout = os.Stdout
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			// Check if the error message, status code, and reason match your criteria
			errorOutput := stderr.String()
			if strings.Contains(errorOutput, "no namespace is found") {
				fmt.Print("You have no project. Please create a new project with kubectl create ns NAMESPACE_NAME")
			} else {
				fmt.Print("kubectl failed:", errorOutput)
			}
			return
		}
		return
	} else {
		if namespace == "default" {
			// Run `kubectl get ns` in the background and get the output
			cmd := exec.Command("kubectl", "get", "ns", "-o", "json")
			var out bytes.Buffer
			cmd.Stdout = &out
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err != nil {
				// Check if the error message, status code, and reason match your criteria
				errorOutput := stderr.String()
				if strings.Contains(errorOutput, "no namespace is found") {
					fmt.Print("You have no project. Please create a new project with kubectl create ns NAMESPACE_NAME")
				} else {
					fmt.Print("kubectl failed:", errorOutput)
				}
				return
			}
			// Parse the JSON output
			var nsList NamespaceList
			err = json.Unmarshal(out.Bytes(), &nsList)
			if err != nil {
				log.Fatalf("Failed to parse JSON: %v", err)
			}

			// Check if the namespace list is empty
			if len(nsList.Items) > 0 {
				fmt.Println("You have projects in CaaS but you need to set in your kubeconfig")
				return
			}
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "Command failed:", err)
				os.Exit(1)
			}
			return
		}

	}

}
