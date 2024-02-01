package main

import (
	"fmt"
	"github.com/koor-tech/genesis/gateway"
	"os"
	"os/exec"
)

func main() {
	r := gateway.SetupRouter()
	r.Run() // listen and serve on 0.0.0.0:8080
}

func examplewithKubeAdm() {
	// initialize a Kubernetes cluster with kubeadm
	initCmd := exec.Command("kubeadm", "init", "--apiserver-advertise-address", "MASTER-NODE-IP")

	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		fmt.Printf("Error executing kubeadm init: %v\n", err)
		return
	}

	fmt.Println("Kubernetes cluster initialized successfully.")

	joinCmd := exec.Command("kubeadm", "join", "MASTER-NODE-I:PORT", "--token", "TOKEN", "--discovery-token-ca-cert-hash", "HASH")

	joinCmd.Stdout = os.Stdout
	joinCmd.Stderr = os.Stderr

	if err := joinCmd.Run(); err != nil {
		fmt.Printf("Error executing kubeadm join: %v\n", err)
		return
	}

	fmt.Println("Node joined to the Kubernetes cluster.")
}
