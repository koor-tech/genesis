package cluster

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/internal/k8s"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/files"
	"github.com/koor-tech/genesis/pkg/genesis"
	"github.com/koor-tech/genesis/pkg/kubeone"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/observer"
	"github.com/koor-tech/genesis/pkg/providers/hetzner"
	"github.com/koor-tech/genesis/pkg/rabbitmq"
	clusters "github.com/koor-tech/genesis/pkg/repositories/postgres/cluster"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/customers"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/providers"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/state"
	"github.com/koor-tech/genesis/pkg/ssh"
	"log"
	"log/slog"
	"os"
	"time"
)

type KoorCluster struct {
	genesisConfig          *genesis.Config
	cluster                *models.Cluster
	subject                *observer.Subject
	templatesSrc           string
	templatesDst           string
	logger                 *slog.Logger
	queue                  *rabbitmq.Client
	clusterStateRepository state.ClusterStateInterface
	customerRepository     customers.CustomerInterface
	clustersRepository     clusters.ClustersInterface
	providerRepository     providers.ProvidersInterface
}

func NewKoorCluster(db *database.DB, rabbitClient *rabbitmq.Client) *KoorCluster {
	_, err := rabbitClient.QueueDeclare()
	if err != nil {
		log.Fatal("unable to declare queue", "error", err)
	}

	kc := KoorCluster{
		subject:                observer.NewSubject(),
		logger:                 slog.New(slog.NewTextHandler(os.Stdout, nil)),
		queue:                  rabbitClient,
		genesisConfig:          genesis.NewConfig(),
		clusterStateRepository: state.NewClusterStateRepository(db),
		customerRepository:     customers.NewCustomersRepository(db),
		providerRepository:     providers.NewProviderRepository(db),
		clustersRepository:     clusters.NewClusterRepository(db),
	}

	// listen the changes in the process/or cluster?
	kc.Listen(context.Background())
	return &kc
}

func (k *KoorCluster) GetCluster(ctx context.Context, ID uuid.UUID) (*models.Cluster, error) {
	c, err := k.clustersRepository.QueryByID(ctx, ID)
	if err != nil {
		k.logger.Error("unable to get cluster", "err", err, "id", ID)
		return nil, err
	}
	return c, nil
}

func (k *KoorCluster) runSshAgent(ctx context.Context, customer *models.Customer) error {
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

	privateKeyFile := fmt.Sprintf("%s/%s", k.getCustomerDir(customer), fileKeyName)
	publicKeyFile := fmt.Sprintf("%s/%s.pub", k.getCustomerDir(customer), fileKeyName)
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

func (k *KoorCluster) getProviderName(provider *models.Provider) string {
	return fmt.Sprintf("%s/%s", k.genesisConfig.TemplatesDir(), provider.Name)
}

func (k *KoorCluster) getCustomerDir(customer *models.Customer) string {
	return fmt.Sprintf("%s/%s", k.genesisConfig.ClientsDir(), customer.ID)
}

func (k *KoorCluster) BuildCluster(ctx context.Context, customer *models.Customer, providerID uuid.UUID) (*models.Cluster, error) {
	k.logger.Info("BuildCluster", "customer", customer.ID)

	customer, err := k.customerRepository.Save(ctx, customer)
	if err != nil {
		k.logger.Error("unable to save client", "err", err)
		return nil, err
	}

	provider, err := k.providerRepository.QueryByID(ctx, providerID)
	if err != nil {
		k.logger.Error("unable to get provider", "err", err)
		return nil, err
	}

	createCluster := models.Cluster{
		ID:       uuid.New(),
		Customer: *customer,
		Provider: *provider,
	}

	providerDir := k.getProviderName(provider)
	customerDir := k.getCustomerDir(customer)

	cluster, err := k.clustersRepository.Save(ctx, createCluster)
	if err != nil {
		k.logger.Error("unable to save the cluster", "err", err)
		return nil, err
	}

	clusterState := models.NewClusterState(cluster)
	clusterState, err = k.clusterStateRepository.Save(ctx, *clusterState)
	if err != nil {
		k.logger.Error("unable to save cluster state", "err", err)
		return cluster, err
	}
	k.logger.Info("Starting")

	k.logger.Info("Setup client environment")
	clusterState.Phase = models.ClusterPhaseSetupInit
	err = k.clusterStateRepository.Update(ctx, *clusterState)
	if err != nil {
		k.logger.Error("unable to save cluster state", "err", err)
		return cluster, err
	}

	// TODO we need to address what will happen the the customer_dir is not empty
	err = files.CopyDir(providerDir, customerDir)
	if err != nil {
		k.logger.Error("unable to copy template files", "err", err)
		return cluster, err
	}

	k.logger.Info("running kubeone")
	kubeOneSvc := kubeone.New(cluster, customerDir)
	k.logger.Info("Creating terraform.tfvars")
	err = kubeOneSvc.WriteTFVars()
	if err != nil {
		k.logger.Error("unable to create terraform.tfvars", "err", err)
		return cluster, err
	}
	k.logger.Info("Generating kubeone.yaml")
	err = kubeOneSvc.WriteConfigFile()
	if err != nil {
		k.logger.Error("unable to create kubeone.yaml", "err", err)
		return cluster, err
	}

	clusterState.Phase = models.ClusterPhaseSetupDone
	err = k.clusterStateRepository.Update(ctx, *clusterState)
	if err != nil {
		k.logger.Error("unable to save cluster state", "err", err)
		return cluster, err
	}

	err = k.subject.Notify(clusterState)
	if err != nil {
		k.logger.Error("unable to notify", "err", err)
		return cluster, err
	}

	k.logger.Info("Done")
	return cluster, nil
}

func (k *KoorCluster) ResumeCluster(ctx context.Context, clusterID uuid.UUID) error {
	k.logger.Info("resume cluster", "id", clusterID)

	c, err := k.GetCluster(ctx, clusterID)
	if err != nil {
		k.logger.Error("unable to get the cluster ", "err", err, "clusterID", clusterID)
	}

	kubeOneSvc := kubeone.New(c, k.getCustomerDir(&c.Customer))
	c.ClusterState.Phase = models.ClusterPhaseSshInit
	clusterState := c.ClusterState

	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}

	k.logger.Info("running ssh agent")
	err = k.runSshAgent(ctx, &c.Customer)
	if err != nil {
		k.logger.Error("unable to run ssh agent", "err", err)
		return err
	}
	clusterState.Phase = models.ClusterPhaseSshDone
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}

	clusterState.Phase = models.ClusterPhaseTerraformInit
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	k.logger.Info("running terraform")
	err = kubeOneSvc.RunTerraform()
	if err != nil {
		k.logger.Error("unable to run terraform", "err", err)
		return err
	}

	clusterState.Phase = models.ClusterPhaseTerraformDone
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	clusterState.Phase = models.ClusterPhaseKubeOneInit
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	k.logger.Info("running kubeone")
	_, err = kubeOneSvc.RunKubeOne()
	if err != nil {
		k.logger.Error("unable to run kubeone", "err", err)
		return err
	}
	clusterState.Phase = models.ClusterPhaseKubeOneDone
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}

	k.logger.Info("Awaiting to get ready the servers 20.. seconds")
	time.Sleep(20 * time.Second)
	clusterState.Phase = models.ClusterPhaseProviderConfInit
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	cloudProvider := hetzner.NewProvider()
	servers, err := cloudProvider.GetServerByLabels(ctx, c.Customer.Name)
	if err != nil {
		k.logger.Error("unable to get server error:", "err", err)
		return err
	}
	k.logger.Info("attaching volumes")
	err = cloudProvider.AttacheVolumesToServers(ctx, c.Customer.Name, servers)
	if err != nil {
		k.logger.Error("unable to get server error:", "err", err)
		return err
	}

	clusterState.Phase = models.ClusterPhaseProviderConfDone
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}

	clusterState.Phase = models.ClusterPhaseClusterReady
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	//
	k.logger.Info("installing rook-ceph")
	kubeConfigName := fmt.Sprintf("koor-client-%s-kubeconfig", c.Customer.Name)
	clusterState.Phase = models.ClusterPhaseInstallCephInit
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	k8sCluster := k8s.New(k.getCustomerDir(&c.Customer) + "/" + kubeConfigName)
	k8sCluster.InstallCharts()
	fmt.Println("======== Installing HELM CHARTS Done! ===========")
	fmt.Println("============ Done ==============")
	clusterState.Phase = models.ClusterPhaseInstallCephDone
	err = k.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		k.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	return nil
}
