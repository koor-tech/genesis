package cluster

import (
	"context"
	"fmt"
	"path"

	"github.com/google/uuid"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/koor-tech/genesis/internal/k8s"
	sshSvc "github.com/koor-tech/genesis/internal/ssh"
	"github.com/koor-tech/genesis/pkg/config"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/kubeone"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/notification"
	"github.com/koor-tech/genesis/pkg/observer"
	"github.com/koor-tech/genesis/pkg/providers/hetzner"
	"github.com/koor-tech/genesis/pkg/rabbitmq"
	clusters "github.com/koor-tech/genesis/pkg/repositories/postgres/cluster"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/customers"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/providers"
	sshRepo "github.com/koor-tech/genesis/pkg/repositories/postgres/ssh"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/state"
	"github.com/koor-tech/genesis/pkg/utils"
	"go.uber.org/fx"
	"go.uber.org/multierr"

	"log/slog"
	"time"
)

type Params struct {
	fx.In

	LC fx.Lifecycle

	Logger       *slog.Logger
	Config       *config.Config
	DB           *database.DB
	RabbitClient *rabbitmq.Client
	Notifier     notification.Notifier
	Provider     *hetzner.Provider
}

type Service struct {
	logger *slog.Logger
	dirCfg config.Directories

	cloudProvider *hetzner.Provider

	subject    *observer.Subject
	sshService *sshSvc.Service

	queue                  *rabbitmq.Client
	clusterStateRepository state.ClusterStateInterface
	customerRepository     customers.CustomerInterface
	clustersRepository     clusters.ClustersInterface
	providerRepository     providers.ProvidersInterface
	sshRepository          sshRepo.SshInterface
	notifier               notification.Notifier
}

func NewService(p Params) (*Service, error) {
	kc := &Service{
		logger:        p.Logger,
		dirCfg:        p.Config.Directories,
		cloudProvider: p.Provider,

		sshService:             sshSvc.NewService(p.Logger, p.DB),
		subject:                observer.NewSubject(),
		queue:                  p.RabbitClient,
		clusterStateRepository: state.NewClusterStateRepository(p.DB),
		customerRepository:     customers.NewCustomersRepository(p.DB),
		providerRepository:     providers.NewProviderRepository(p.DB),
		clustersRepository:     clusters.NewClusterRepository(p.DB),
		sshRepository:          sshRepo.NewSshRepository(p.DB),
		notifier:               p.Notifier,
	}

	p.LC.Append(fx.StartHook(func(ctx context.Context) error {
		_, err := p.RabbitClient.QueueDeclare()
		if err != nil {
			return fmt.Errorf("unable to declare queue. %w", err)
		}

		return nil
	}))

	// Listen the changes in the process/or cluster?
	kc.Listen(context.Background())

	return kc, nil
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
	return path.Join(s.dirCfg.TemplatesDir(), provider.Name)
}

func (s *Service) getCustomerDir(customer *models.Customer) string {
	return path.Join(s.dirCfg.ClientsDir(), customer.ID.String())
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
	err = utils.CopyDir(providerDir, customerDir)
	if err != nil {
		s.logger.Error("unable to copy template files", "err", err)
		return cluster, err
	}

	s.logger.Info("running kubeone")
	kubeOneSvc, err := kubeone.New(s.logger, cluster,
		customerDir, s.cloudProvider)
	if err != nil {
		return cluster, err
	}

	s.logger.Info("Creating terraform.tfvars")
	err = kubeOneSvc.WriteTFVars()
	if err != nil {
		s.logger.Error("unable to create terraform.tfvars", "err", err)
		return cluster, err
	}
	s.logger.Info("Generating kubeone.yaml")
	err = kubeOneSvc.WriteConfigFile(cluster.ID.String())
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

	cluster, err = s.GetCluster(ctx, cluster.ID)
	if err != nil {
		s.logger.Error("unable to find the cluster", "err", err)
		return nil, err
	}
	s.logger.Info("Done")
	return cluster, nil
}

func (s *Service) ResumeCluster(ctx context.Context, clusterID uuid.UUID) error {
	s.logger.Info("resume cluster", "id", clusterID)

	cluster, err := s.GetCluster(ctx, clusterID)
	if err != nil {
		s.logger.Error("unable to get the cluster ", "err", err, "clusterID", clusterID)
	}

	kubeOneSvc, err := kubeone.New(s.logger, cluster,
		s.getCustomerDir(&cluster.Customer), s.cloudProvider)
	if err != nil {
		return err
	}

	cluster.ClusterState.Phase = models.ClusterPhaseSshInit
	clusterState := cluster.ClusterState

	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}

	s.logger.Info("running ssh agent")
	_, err = s.sshService.BuildAndRunSSH(ctx, clusterID, s.getCustomerDir(&cluster.Customer))
	if err != nil {
		s.logger.Error("unable to run ssh agent", "err", err)
		return err
	}
	clusterState.Phase = models.ClusterPhaseSshDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}

	clusterState.Phase = models.ClusterPhaseTerraformInit
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}
	s.logger.Info("running terraform")
	err = kubeOneSvc.RunTerraform(ctx)
	if err != nil {
		s.logger.Error("unable to run terraform", "err", err)
		return err
	}

	clusterState.Phase = models.ClusterPhaseTerraformDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}
	clusterState.Phase = models.ClusterPhaseKubeOneInit
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
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
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}

	s.logger.Info("Awaiting to get ready the servers 120.. seconds")
	time.Sleep(120 * time.Second)
	clusterState.Phase = models.ClusterPhaseProviderConfInit
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}

	servers, err := s.cloudProvider.GetServerByLabels(ctx, cluster.Customer.Name)
	if err != nil {
		s.logger.Error("unable to get server error:", "err", err)
		return err
	}
	s.logger.Info("attaching volumes")
	err = s.cloudProvider.AttacheVolumesToServers(ctx, cluster.Customer.Name, servers)
	if err != nil {
		s.logger.Error("unable to get server error:", "err", err)
		return err
	}

	clusterState.Phase = models.ClusterPhaseProviderConfDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}

	clusterState.Phase = models.ClusterPhaseClusterReady
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}
	//
	s.logger.Info("installing rook-ceph")
	kubeConfigName := fmt.Sprintf("koor-client-%s-kubeconfig", cluster.Customer.Name)
	clusterState.Phase = models.ClusterPhaseInstallCephInit
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}

	kubeConfigFile := s.getCustomerDir(&cluster.Customer) + "/" + kubeConfigName

	k8sCluster := k8s.New(kubeConfigFile, s.dirCfg.ChartsDir())
	err = k8sCluster.InstallCharts()
	if err != nil {
		s.logger.Error("unable to install rook-ceph charts ", "err ", err, "clusterID", cluster.ID)
		return err
	}
	clusterState.Phase = models.ClusterPhaseInstallCephDone
	err = s.clusterStateRepository.Update(ctx, clusterState)
	if err != nil {
		s.logger.Error("unable to save the state of the cluster ", "err ", err, "clusterID", cluster.ID)
	}

	s.logger.Info("Ceph Cluster provisioned")

	kubeConfigData, err := utils.ReadFileAsString(kubeConfigFile)
	if err != nil {
		s.logger.Error("unable to get the content of kubeConfig file", "file", kubeConfigFile, "err", err)
		return err
	}

	cluster.KubeConfig = &kubeConfigData
	err = s.clustersRepository.Update(ctx, *cluster)
	if err != nil {
		s.logger.Error("unable to update the cluster ", "err ", err, "clusterID", cluster.ID)
	}

	if err := s.notifier.Send(cluster.Customer); err != nil {
		s.logger.Error("failed to notify customer", "err", err)
	}

	return nil
}

func (s *Service) DeleteCluster(ctx context.Context, clusterID uuid.UUID) error {
	clusterLabel := utils.Label("kubeone_cluster_name", clusterID.String())
	listOpts := hcloud.ListOpts{
		LabelSelector: clusterLabel,
	}

	errs := multierr.Combine()

	// Servers
	servers, err := s.cloudProvider.Client.Server.AllWithOpts(ctx, hcloud.ServerListOpts{
		ListOpts: listOpts,
	})
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	for _, item := range servers {
		_, _, err := s.cloudProvider.Client.Server.DeleteWithResult(ctx, item)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}

	// Networks
	networks, err := s.cloudProvider.Client.Network.AllWithOpts(ctx, hcloud.NetworkListOpts{
		ListOpts: listOpts,
	})
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	for _, item := range networks {
		_, err := s.cloudProvider.Client.Network.Delete(ctx, item)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}

	// Load balancer
	lbs, err := s.cloudProvider.Client.LoadBalancer.AllWithOpts(ctx, hcloud.LoadBalancerListOpts{
		ListOpts: listOpts,
	})
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	for _, item := range lbs {
		_, err := s.cloudProvider.Client.LoadBalancer.Delete(ctx, item)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}

	// Firewalls
	firewalls, err := s.cloudProvider.Client.Firewall.AllWithOpts(ctx, hcloud.FirewallListOpts{
		ListOpts: listOpts,
	})
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	for _, item := range firewalls {
		_, err := s.cloudProvider.Client.Firewall.Delete(ctx, item)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}

	// Volumes
	volumes, err := s.cloudProvider.Client.Volume.AllWithOpts(ctx, hcloud.VolumeListOpts{
		ListOpts: listOpts,
	})
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	for _, item := range volumes {
		_, err := s.cloudProvider.Client.Volume.Delete(ctx, item)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}

	// Every time we create an additional resource, we need to add it here

	return errs
}
