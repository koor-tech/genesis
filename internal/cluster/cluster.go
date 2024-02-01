package cluster

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/internal/k8s"
	"github.com/koor-tech/genesis/pkg/files"
	"github.com/koor-tech/genesis/pkg/kubeone"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/providers/hetzner"
	"github.com/koor-tech/genesis/pkg/ssh"
	"log/slog"
	"os"
	"time"
)

// this is the good one for docker
// const koorClientsDir = "/koor/clients"
// temp for dev
const koorClientsDir = "/home/javier/koor"

type KoorCluster struct {
	cluster      *models.Cluster
	templatesSrc string
	templatesDst string
	logger       *slog.Logger
}

func NewKoorCluster(provider string, client *models.Client) *KoorCluster {

	return &KoorCluster{
		cluster: &models.Cluster{
			ID:       uuid.New(),
			Provider: provider,
			Client:   client,
		},
		templatesSrc: koorClientsDir + "/templates/" + provider,
		templatesDst: koorClientsDir + "/clients/" + client.ID.String(),
		logger:       slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (k *KoorCluster) Cluster() models.Cluster {
	return *k.cluster
}

func (k *KoorCluster) BuildCluster(ctx context.Context) error {
	k.logger.Info("BuildCluster", "Provider", k.cluster.Provider, "ID", k.cluster.ID, "Client", k.cluster.Client.ID)

	fmt.Println("============ Starting  ==============")

	fmt.Println("============ Setup client environment  ==============")
	// copy files to keep a copy fo the resources required by kubeone
	err := files.CopyDir(k.templatesSrc, k.templatesDst)
	if err != nil {
		k.logger.Error("unable to copy template files", "err", err)
		return err
	}

	fmt.Println("============ Generating SSH Keys  ==============")
	//Generate SSH key
	keys, err := ssh.GenerateKey()
	if err != nil {
		k.logger.Error("unable to generate ssh keys", "err", err)
		return err
	}
	fmt.Println("============ Saving SSH Keys  ==============")
	//Save ssh into clients folder
	err = files.SaveInFile(k.templatesDst+"/id_ed25519", keys.Private, 0600)
	if err != nil {
		k.logger.Error("unable to generate ssh keys", "err", err)
		return err
	}
	err = files.SaveInFile(k.templatesDst+"/id_ed25519.pub", keys.Public, 0600)
	if err != nil {
		k.logger.Error("unable to generate ssh keys", "err", err)
		return err
	}

	fmt.Println("============ Running SSH agent  ==============")
	err = ssh.RunAgent(k.templatesDst + "/id_ed25519")
	if err != nil {
		k.logger.Error("unable to execute ssh agent", "err", err)
		return err
	}
	fmt.Println("============ Running Kubeone service  ==============")
	kubeOneSvc := kubeone.New(k.cluster, k.templatesSrc, k.templatesDst)
	fmt.Println("============ Generating terraform.tfvars  ==============")
	err = kubeOneSvc.WriteTFVars()
	if err != nil {
		k.logger.Error("unable to create terraform.tfvars", "err", err)
		return err
	}
	fmt.Println("============ Generating kubeone.yaml  ==============")
	err = kubeOneSvc.WriteConfigFile()
	if err != nil {
		k.logger.Error("unable to create kubeone.yaml", "err", err)
		return err
	}
	fmt.Println("============ running terraform  ==============")
	err = kubeOneSvc.RunTerraform()
	if err != nil {
		k.logger.Error("unable to run terraform", "err", err)
		return err
	}
	fmt.Println("============ running kubeone  ==============")
	_, err = kubeOneSvc.RunKubeOne()
	if err != nil {
		k.logger.Error("unable to run kubeone", "err", err)
		return err
	}

	fmt.Println("----- Awaiting to get ready the servers 60.. seconds -----------")
	time.Sleep(60 * time.Second)

	token := "..."
	cloudProvider := hetzner.NewProvider(token)
	fmt.Println("======== Attaching Volumes! ===========")
	labels := map[string]string{
		"kubeone_cluster_name": fmt.Sprintf("koor-client-%s", k.cluster.Client.Name),
		fmt.Sprintf("koor-client-%s-workers", k.cluster.Client.Name): "pool1",
	}
	//	labels := fmt.Sprintf("kubeone_cluster_name=koor-client-%s,koor-client-%s-workers=pool1", customer, customer)
	servers, err := cloudProvider.GetServerByLabels(ctx, labels)
	if err != nil {
		k.logger.Error("unable to get server error:", "err", err)
		return err
	}

	volLabels := map[string]string{"koor-client": k.cluster.Client.Name, "pool": "pool1"}
	for _, server := range servers {
		serverName := server.Name
		fmt.Println("======== Attaching Volume to " + serverName + " ===========")
		name := fmt.Sprintf("data-%s-pool1-%s", k.cluster.Client.Name, serverName[len(serverName)-5:])
		err = cloudProvider.AttachVolumeToServer(ctx, server, volLabels, 40, name, false)
		if err != nil {
			k.logger.Error("unable to create volume:", "err", err)
			return err
		}
	}

	fmt.Println("======== Attaching Volumes Done! ===========")
	fmt.Println("======== Installing HELM CHARTS ===========")

	kubeConfigName := fmt.Sprintf("koor-client-%s-kubeconfig", k.cluster.Client.Name)
	k8sCluster := k8s.New(k.templatesDst + "/" + kubeConfigName)
	k8sCluster.InstallCharts()
	fmt.Println("======== Installing HELM CHARTS Done! ===========")
	fmt.Println("============ Done ==============")
	return nil
}
