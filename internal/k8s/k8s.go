package k8s

import (
	"context"
	"fmt"
	"os"
	"path"

	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	rookCephClusterChartValues  = "rook-ceph-cluster/values.yaml"
	rookCephOperatorChartValues = "rook-ceph/values.yaml"
)

const (
	defaultRookCephNS = "rook-ceph"
)

type Cluster struct {
	kubeConfigFile string
	chartsDir      string
}

func New(kubeConfigFile string, chartsDir string) *Cluster {
	return &Cluster{
		kubeConfigFile: kubeConfigFile,
		chartsDir:      chartsDir,
	}
}

func (c *Cluster) InstallCharts() error {
	fmt.Println("======== Installing helm charts running... ===========")
	// read configurations
	kubeConf, err := os.ReadFile(c.kubeConfigFile)
	if err != nil {
		return fmt.Errorf("could not load KubeConfig. %w", err)
	}

	// rook-ceph operator values
	rookCephOperatorValues := path.Join(c.chartsDir, rookCephOperatorChartValues)
	valuesRookOperatorYaml, err := os.ReadFile(rookCephOperatorValues)
	if err != nil {
		return fmt.Errorf("could not load values.yaml for rook-operator: %w", err)
	}

	// rook-ceph cluster values
	rookCephClusterValues := path.Join(c.chartsDir, rookCephClusterChartValues)
	valuesRookClusterYaml, err := os.ReadFile(rookCephClusterValues)
	if err != nil {
		return fmt.Errorf("could not load values.yaml for rook-cluster: %w", err)
	}

	fmt.Println("======== Preparing helm .. ===========")

	opt := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        defaultRookCephNS,
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true,
			DebugLog: func(format string, v ...interface{}) {
			},
		},
		KubeConfig: kubeConf,
	}

	helmClient, err := helmclient.NewClientFromKubeConf(opt, helmclient.Burst(100), helmclient.Timeout(10e9))
	if err != nil {
		return err
	}

	chartRepo := repo.Entry{
		Name: "rook-release",
		URL:  "https://charts.rook.io/release",
	}

	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		return err
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName:     "rook-ceph",
		ChartName:       "rook-release/rook-ceph",
		Namespace:       defaultRookCephNS,
		CreateNamespace: true,
		UpgradeCRDs:     true,
		ValuesYaml:      string(valuesRookOperatorYaml),
	}

	fmt.Println("========== installing Ceph Operator Helm Chart =======")

	release, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		return err
	}

	fmt.Println("========= Ceph Operator Deployed =================")
	fmt.Printf("release: %+v\n", release.Name)

	//helm install --create-namespace --namespace rook-ceph rook-ceph-cluster \
	//--set operatorNamespace=rook-ceph rook-release/rook-ceph-cluster -f values.yaml
	chartClusterSpec := &helmclient.ChartSpec{
		ReleaseName:     "rook-ceph-cluster",
		ChartName:       "rook-release/rook-ceph-cluster",
		Namespace:       defaultRookCephNS,
		ValuesYaml:      string(valuesRookClusterYaml),
		CreateNamespace: true,
		Force:           true,
	}

	fmt.Println("========== installing Ceph GetCluster Helm Chart  =======")
	clusterRelease, err := helmClient.InstallOrUpgradeChart(context.Background(), chartClusterSpec, nil)
	if err != nil {
		return err
	}
	fmt.Println("========= Ceph GetCluster Deployed =================")
	fmt.Printf("release: %+v\n", clusterRelease.Name)

	releases, err := helmClient.ListDeployedReleases()
	if err != nil {
		return err
	}

	fmt.Println("========= List of charts installed =================")

	for _, release := range releases {
		fmt.Printf("name: %s notes: %s\n", release.Name, release.Info.Notes)
	}
	fmt.Println("================= CHARTS installed!=========")

	return nil
}
