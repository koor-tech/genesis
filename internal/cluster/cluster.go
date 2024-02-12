package cluster

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/gateway/request"
	"github.com/koor-tech/genesis/internal/k8s"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/files"
	"github.com/koor-tech/genesis/pkg/kubeone"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/observer"
	"github.com/koor-tech/genesis/pkg/providers/hetzner"
	"github.com/koor-tech/genesis/pkg/rabbitmq"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/clients"
	clusters "github.com/koor-tech/genesis/pkg/repositories/postgres/cluster"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/providers"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/state"
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
	cluster                *models.Cluster
	subject                *observer.Subject
	templatesSrc           string
	templatesDst           string
	logger                 *slog.Logger
	queue                  *rabbitmq.Client
	clusterStateRepository state.ClusterStateInterface
	clientsRepository      clients.ClientsInterface
	clustersRepository     clusters.ClustersInterface
	providerRepository     providers.ProvidersInterface
}

func NewKoorCluster(db *database.DB) *KoorCluster {
	queue := rabbitmq.NewClient()
	queue.QueueDeclare()
	kc := KoorCluster{
		subject:                observer.NewSubject(),
		logger:                 slog.New(slog.NewTextHandler(os.Stdout, nil)),
		queue:                  queue,
		templatesSrc:           koorClientsDir + "/templates/",
		templatesDst:           koorClientsDir + "/clients/",
		clusterStateRepository: state.NewClusterStateRepository(db),
		clientsRepository:      clients.NewClientsRepository(db),
		providerRepository:     providers.NewProviderRepository(db),
		clustersRepository:     clusters.NewClusterRepository(db),
	}

	kc.Listen(context.Background())
	return &kc
}

func (k *KoorCluster) NewCluster(ctx context.Context, params request.CreateClusterRequest) error {
	createClient := models.Client{
		ID:   uuid.New(),
		Name: params.ClientName,
	}
	client, err := k.clientsRepository.Save(ctx, createClient)
	if err != nil {
		k.logger.Error("unable to save client", "err", err)
		return err
	}

	provider, err := k.providerRepository.QueryByID(ctx, uuid.MustParse("80be226b-8355-4dea-b41a-6e17ea37559a"))
	if err != nil {
		k.logger.Error("unable to get provider", "err", err)
		return err
	}

	createCluster := models.Cluster{
		ID:       uuid.New(),
		Client:   *client,
		Provider: *provider,
	}

	cluster, err := k.clustersRepository.Save(ctx, createCluster)
	if err != nil {
		k.logger.Error("unable to save the cluster", "err", err)
		return err
	}

	k.templatesSrc = k.templatesSrc + cluster.Provider.Name
	k.templatesDst = k.templatesDst + cluster.Client.ID.String()
	k.cluster = cluster
	return nil
}

func (k *KoorCluster) Cluster(ctx context.Context, ID uuid.UUID) (*models.Cluster, error) {
	c, err := k.clustersRepository.QueryByID(ctx, ID)
	fmt.Println("================================")
	fmt.Printf("ID: %+v\n", ID)
	if err != nil {
		k.logger.Error("unable to get cluster", "err", err, "id", ID)
		return nil, err
	}
	return c, nil
}

func (k *KoorCluster) runSshAgent(ctx context.Context) error {
	const (
		filePermission = 0600
		fileKeyName    = "id_ed25519"
	)
	//Generate SSH key
	keys, err := ssh.GenerateKey()
	if err != nil {
		k.logger.Error("unable to generate ssh keys", "err", err)
		return err
	}

	privateKeyFile := fmt.Sprintf("%s/%s", k.templatesDst, fileKeyName)
	publicKeyFile := fmt.Sprintf("%s/%s.pub", k.templatesDst, fileKeyName)
	//Save ssh into clients folder
	err = files.SaveInFile(privateKeyFile, keys.Private, filePermission)
	if err != nil {
		k.logger.Error("unable to generate ssh keys", "err", err)
		return err
	}
	err = files.SaveInFile(publicKeyFile, keys.Public, filePermission)
	if err != nil {
		k.logger.Error("unable to generate ssh keys", "err", err)
		return err
	}

	k.logger.Info("============ Running SSH agent  ==============")
	err = ssh.RunAgent(privateKeyFile)
	if err != nil {
		k.logger.Error("unable to execute ssh agent", "err", err)
		return err
	}
	return nil
}

func (k *KoorCluster) BuildCluster(ctx context.Context) (*models.ClusterState, error) {
	k.logger.Info("BuildCluster", "Provider", k.cluster.Provider.Name, "ID", k.cluster.ID, "Client", k.cluster.Client.ID)

	clusterState := models.NewClusterState(k.cluster)
	clusterState, err := k.clusterStateRepository.Save(ctx, *clusterState)
	if err != nil {
		k.logger.Error("unable to save cluster state", "err", err)
		return clusterState, err
	}
	k.logger.Info("Starting")

	k.logger.Info("Setup client environment")
	clusterState.Phase = models.ClusterPhaseSetupInit
	err = k.clusterStateRepository.Update(ctx, *clusterState)
	if err != nil {
		k.logger.Error("unable to save cluster state", "err", err)
		return clusterState, err
	}

	err = files.CopyDir(k.templatesSrc, k.templatesDst)
	if err != nil {
		k.logger.Error("unable to copy template files", "err", err)
		return clusterState, err
	}

	k.logger.Info("running kubeone")
	kubeOneSvc := kubeone.New(k.cluster, k.templatesSrc, k.templatesDst)
	k.logger.Info("Creating terraform.tfvars")
	err = kubeOneSvc.WriteTFVars()
	if err != nil {
		k.logger.Error("unable to create terraform.tfvars", "err", err)
		return clusterState, err
	}
	k.logger.Info("Generating kubeone.yaml")
	err = kubeOneSvc.WriteConfigFile()
	if err != nil {
		k.logger.Error("unable to create kubeone.yaml", "err", err)
		return clusterState, err
	}

	clusterState.Phase = models.ClusterPhaseSetupDone
	err = k.clusterStateRepository.Update(ctx, *clusterState)
	if err != nil {
		k.logger.Error("unable to save cluster state", "err", err)
		return clusterState, err
	}

	err = k.subject.Notify(clusterState)
	if err != nil {
		k.logger.Error("unable to notify", "err", err)
		return clusterState, err
	}

	k.logger.Info("Done")
	return clusterState, nil
}

func (k *KoorCluster) ResumeCluster(ctx context.Context, clusterState models.ClusterState) error {
	k.logger.Info("resume cluster", "id", clusterState.ID)

	c, err := k.Cluster(ctx, clusterState.ClusterID)
	if err != nil {
		k.logger.Error("unable to get the cluster ", "err", err, "clusterID", clusterState.ID)
	}

	fmt.Println("============ Resume Cluster state  ==============")
	k.templatesSrc = k.templatesSrc + "hetzner"
	k.templatesDst = k.templatesDst + c.Client.ID.String()
	kubeOneSvc := kubeone.New(c, k.templatesSrc, k.templatesDst)
	c.ClusterState.Phase = models.ClusterPhaseSshInit
	cs := c.ClusterState

	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}

	k.logger.Info("running ssh agent")
	err = k.runSshAgent(ctx)
	if err != nil {
		k.logger.Error("unable to run ssh agent", "err", err)
		return err
	}
	cs.Phase = models.ClusterPhaseSshDone
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}

	cs.Phase = models.ClusterPhaseTerraformInit
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}
	k.logger.Info("running terraform")
	err = kubeOneSvc.RunTerraform()
	if err != nil {
		k.logger.Error("unable to run terraform", "err", err)
		return err
	}

	cs.Phase = models.ClusterPhaseTerraformDone
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}
	cs.Phase = models.ClusterPhaseKubeOneInit
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}
	k.logger.Info("running kubeone")
	_, err = kubeOneSvc.RunKubeOne()
	if err != nil {
		k.logger.Error("unable to run kubeone", "err", err)
		return err
	}
	cs.Phase = models.ClusterPhaseKubeOneDone
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}

	k.logger.Info("Awaiting to get ready the servers 20.. seconds")
	time.Sleep(20 * time.Second)
	cs.Phase = models.ClusterPhaseProviderConfInit
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}
	cloudProvider := hetzner.NewProvider(c.Client.Name)
	servers, err := cloudProvider.GetServerByLabels(ctx)
	if err != nil {
		k.logger.Error("unable to get server error:", "err", err)
		return err
	}
	k.logger.Info("attaching volumes")
	err = cloudProvider.AttacheVolumesToServers(ctx, servers)
	if err != nil {
		k.logger.Error("unable to get server error:", "err", err)
		return err
	}

	cs.Phase = models.ClusterPhaseProviderConfDone
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}

	cs.Phase = models.ClusterPhaseClusterReady
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}
	//
	k.logger.Info("installing rook-ceph")
	kubeConfigName := fmt.Sprintf("koor-client-%s-kubeconfig", c.Client.Name)
	cs.Phase = models.ClusterPhaseInstallCephInit
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}
	k8sCluster := k8s.New(k.templatesDst + "/" + kubeConfigName)
	k8sCluster.InstallCharts()
	fmt.Println("======== Installing HELM CHARTS Done! ===========")
	fmt.Println("============ Done ==============")
	cs.Phase = models.ClusterPhaseInstallCephDone
	err = k.clusterStateRepository.Update(ctx, cs)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", clusterState.ID)
	}
	return nil
}

//
//func (k *KoorCluster) GetKoorCluster(ctx context.Context, id uuid.UUID) (*models.Cluster, error) {
//
//	return nil, nil
//}
//
//func (k *KoorCluster) RemoveResources(ctx context.Context) {
//	// Getting volumes
//	// Dettach volumes
//	//run terraform destroy
//}
