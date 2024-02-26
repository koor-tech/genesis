package cluster

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/internal/k8s"
	sshSvc "github.com/koor-tech/genesis/internal/ssh"
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
	sshRepo "github.com/koor-tech/genesis/pkg/repositories/postgres/ssh"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/state"

	"log"
	"log/slog"
	"os"
	"time"
)

type Service struct {
	genesisConfig *genesis.Config
	cluster       *models.Cluster
	subject       *observer.Subject
	sshService    *sshSvc.Service

	templatesSrc           string
	templatesDst           string
	logger                 *slog.Logger
	queue                  *rabbitmq.Client
	clusterStateRepository state.ClusterStateInterface
	customerRepository     customers.CustomerInterface
	clustersRepository     clusters.ClustersInterface
	providerRepository     providers.ProvidersInterface
	sshRepository          sshRepo.SshInterface
}

func NewService(db *database.DB, rabbitClient *rabbitmq.Client) *Service {
	_, err := rabbitClient.QueueDeclare()
	if err != nil {
		log.Fatal("unable to declare queue", "error", err)
	}

	kc := Service{
		sshService:             sshSvc.NewService(db),
		subject:                observer.NewSubject(),
		logger:                 slog.New(slog.NewTextHandler(os.Stdout, nil)),
		queue:                  rabbitClient,
		genesisConfig:          genesis.NewConfig(),
		clusterStateRepository: state.NewClusterStateRepository(db),
		customerRepository:     customers.NewCustomersRepository(db),
		providerRepository:     providers.NewProviderRepository(db),
		clustersRepository:     clusters.NewClusterRepository(db),
		sshRepository:          sshRepo.NewSshRepository(db),
	}

	// listen the changes in the process/or cluster?
	kc.Listen(context.Background())
	return &kc
}

func (s *Service) GetCluster(ctx context.Context, ID uuid.UUID) (*models.Cluster, error) {
	c, err := s.clustersRepository.QueryByID(ctx, ID)
	if err != nil {
		s.logger.Error("unable to get cluster", "err", err, "id", ID)
		return nil, err
	}
	return c, nil
}

func (s *Service) getProviderName(provider *models.Provider) string {
	return fmt.Sprintf("%s/%s", s.genesisConfig.TemplatesDir(), provider.Name)
}

func (s *Service) getCustomerDir(customer *models.Customer) string {
	return fmt.Sprintf("%s/%s", s.genesisConfig.ClientsDir(), customer.ID)
}

func (s *Service) BuildCluster(ctx context.Context, customer *models.Customer, providerID uuid.UUID) (*models.Cluster, error) {
	s.logger.Info("BuildCluster", "customer", customer.ID)

	customer, err := s.customerRepository.Save(ctx, customer)
	if err != nil {
		s.logger.Error("unable to save client", "err", err)
		return nil, err
	}

	provider, err := s.providerRepository.QueryByID(ctx, providerID)
	if err != nil {
		s.logger.Error("unable to get provider", "err", err)
		return nil, err
	}

	createCluster := models.Cluster{
		ID:       uuid.New(),
		Customer: *customer,
		Provider: *provider,
	}

	providerDir := s.getProviderName(provider)
	customerDir := s.getCustomerDir(customer)

	cluster, err := s.clustersRepository.Save(ctx, createCluster)
	if err != nil {
		s.logger.Error("unable to save the cluster", "err", err)
		return nil, err
	}

	clusterState := models.NewClusterState(cluster)
	clusterState, err = s.clusterStateRepository.Save(ctx, *clusterState)
	if err != nil {
		s.logger.Error("unable to save cluster state", "err", err)
		return cluster, err
	}
	s.logger.Info("Starting")

	s.logger.Info("Setup client environment")
	clusterState.Phase = models.ClusterPhaseSetupInit
	err = s.clusterStateRepository.Update(ctx, *clusterState)
	if err != nil {
		s.logger.Error("unable to save cluster state", "err", err)
		return cluster, err
	}

	// TODO we need to address what will happen the the customer_dir is not empty
	err = files.CopyDir(providerDir, customerDir)
	if err != nil {
		s.logger.Error("unable to copy template files", "err", err)
		return cluster, err
	}

	s.logger.Info("running kubeone")
	kubeOneSvc := kubeone.New(cluster, customerDir)
	s.logger.Info("Creating terraform.tfvars")
	err = kubeOneSvc.WriteTFVars()
	if err != nil {
		s.logger.Error("unable to create terraform.tfvars", "err", err)
		return cluster, err
	}
	s.logger.Info("Generating kubeone.yaml")
	err = kubeOneSvc.WriteConfigFile()
	if err != nil {
		s.logger.Error("unable to create kubeone.yaml", "err", err)
		return cluster, err
	}

	clusterState.Phase = models.ClusterPhaseSetupDone
	err = s.clusterStateRepository.Update(ctx, *clusterState)
	if err != nil {
		s.logger.Error("unable to save cluster state", "err", err)
		return cluster, err
	}

	err = s.subject.Notify(clusterState)
	if err != nil {
		s.logger.Error("unable to notify", "err", err)
		return cluster, err
	}

	s.logger.Info("Done")
	return cluster, nil
}

func (s *Service) ResumeCluster(ctx context.Context, clusterID uuid.UUID) error {
	s.logger.Info("resume cluster", "id", clusterID)

	c, err := s.GetCluster(ctx, clusterID)
	if err != nil {
		s.logger.Error("unable to get the cluster ", "err", err, "clusterID", clusterID)
	}

	kubeOneSvc := kubeone.New(c, s.getCustomerDir(&c.Customer))
	c.ClusterState.Phase = models.ClusterPhaseSshInit
	clusterState := c.ClusterState

	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}

	s.logger.Info("running ssh agent")
	_, err = s.sshService.BuildAndRunSSH(ctx, clusterID, s.getCustomerDir(&c.Customer))
	if err != nil {
		s.logger.Error("unable to run ssh agent", "err", err)
		return err
	}
	clusterState.Phase = models.ClusterPhaseSshDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}

	clusterState.Phase = models.ClusterPhaseTerraformInit
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	s.logger.Info("running terraform")
	err = kubeOneSvc.RunTerraform()
	if err != nil {
		s.logger.Error("unable to run terraform", "err", err)
		return err
	}

	clusterState.Phase = models.ClusterPhaseTerraformDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	clusterState.Phase = models.ClusterPhaseKubeOneInit
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	s.logger.Info("running kubeone")
	_, err = kubeOneSvc.RunKubeOne()
	if err != nil {
		s.logger.Error("unable to run kubeone", "err", err)
		return err
	}
	clusterState.Phase = models.ClusterPhaseKubeOneDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}

	s.logger.Info("Awaiting to get ready the servers 20.. seconds")
	time.Sleep(20 * time.Second)
	clusterState.Phase = models.ClusterPhaseProviderConfInit
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	cloudProvider := hetzner.NewProvider()
	servers, err := cloudProvider.GetServerByLabels(ctx, c.Customer.Name)
	if err != nil {
		s.logger.Error("unable to get server error:", "err", err)
		return err
	}
	s.logger.Info("attaching volumes")
	err = cloudProvider.AttacheVolumesToServers(ctx, c.Customer.Name, servers)
	if err != nil {
		s.logger.Error("unable to get server error:", "err", err)
		return err
	}

	clusterState.Phase = models.ClusterPhaseProviderConfDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}

	clusterState.Phase = models.ClusterPhaseClusterReady
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	//
	s.logger.Info("installing rook-ceph")
	kubeConfigName := fmt.Sprintf("koor-client-%s-kubeconfig", c.Customer.Name)
	clusterState.Phase = models.ClusterPhaseInstallCephInit
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	k8sCluster := k8s.New(s.getCustomerDir(&c.Customer) + "/" + kubeConfigName)
	k8sCluster.InstallCharts()
	clusterState.Phase = models.ClusterPhaseInstallCephDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", c.ID)
	}
	s.logger.Info("Ceph Cluster provisioned")
	return nil
}
