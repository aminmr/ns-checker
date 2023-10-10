package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"k8s.io/client-go/tools/clientcmd"
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

	// Create Kubernetes client
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
	flag.Parse()
	args := flag.Args()
	cmd := exec.Command("kubectl", args...)

	// Check if the command is 'kubectl create ns' or 'kubectl get ns', if so, ignore
	if len(args) > 1 && ((args[0] == "create" && args[1] == "ns") || (args[0] == "get" && args[1] == "ns")) {
		// Execute the original kubectl command
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Command failed:", err)
			os.Exit(1)
		}
		return
	} else {
		if namespace == "default" {
			// Run `kubectl get ns` in the background and get the output
			cmd := exec.Command("kubectl", "get", "ns", "-o", "json")
			var out bytes.Buffer
			cmd.Stdout = &out
			err := cmd.Run()
			if err != nil {
				log.Fatalf("Failed to get namespaces: %v", err)
			}

			// Parse the JSON output
			var nsList NamespaceList
			err = json.Unmarshal(out.Bytes(), &nsList)
			if err != nil {
				log.Fatalf("Failed to parse JSON: %v", err)
			}

			// Check if the namespace list is empty
			if len(nsList.Items) == 0 {
				fmt.Println("You have no project. Please create a new project with kubectl create ns NAMESPACE_NAME")
				return
			} else {
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
