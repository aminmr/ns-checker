/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "A wrapper for kubectl to check namespace existence",
	Run: func(cmd *cobra.Command, args []string) {
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

		// Get namespace from kubeconfig
		namespace, _, err := configLoader.Namespace()
		if err != nil {
			fmt.Println("Error retrieving namespace:", err)
			return
		}

		// Create Kubernetes client
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Println("Error getting kubeconfig:", err)
			return
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Println("Error creating Kubernetes client:", err)
			return
		}

		// Check if the command is 'kubectl create ns' or 'kubectl get ns', if so, ignore
		if len(args) > 1 && ((args[0] == "create" && args[1] == "ns") || (args[0] == "get" && args[1] == "ns")) {
			// Execute the original kubectl command
			executeKubectl(args)
			return
		}

		// Check if namespace is set in kubeconfig context
		if namespace == "" {
			// List namespaces
			namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Println("Error listing namespaces:", err)
				return
			}

			if len(namespaces.Items) == 0 {
				fmt.Println("You have no namespaces. Please create a new one using 'kubectl create ns <namespace>'.")
			} else {
				fmt.Println("You have namespaces but none is set in kubeconfig. Please set a namespace using 'kubectl config set-context --current --namespace=<namespace>'.")
			}
			return
		}

		// Check if namespace exists in the cluster
		_, err = clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
		if err != nil {
			fmt.Println("Namespace does not exist in the cluster. Please create it using 'kubectl create ns", namespace, "'.")
			return
		}

		// If namespace exists, proceed with the command silently
		executeKubectl(args)
	},
}


func executeKubectl(args []string) {
	// Execute the original kubectl command
	kubectlCmd := exec.Command("/usr/local/bin/kubectl", args...)
	kubectlCmd.Stdout = os.Stdout
	kubectlCmd.Stderr = os.Stderr
	kubectlCmd.Stdin = os.Stdin

	err := kubectlCmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error executing kubectl command:", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// func init() {
// 	// Here you will define your flags and configuration settings.
// 	// Cobra supports persistent flags, which, if defined here,
// 	// will be global for your application.

// 	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubectl-plugin.yaml)")

// 	// Cobra also supports local flags, which will only run
// 	// when this action is called directly.
// 	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
