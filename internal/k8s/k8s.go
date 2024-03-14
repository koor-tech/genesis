package k8s

import (
	"context"
	"fmt"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"os"
)

const (
	rookCephClusterChartValues  = "charts/rook-ceph-cluster/values.yaml"
	rookCephOperatorChartValues = "charts/rook-ceph/values.yaml"
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

func (c *Cluster) InstallCharts() {
	fmt.Println("======== Installing helm charts running... ===========")
	// read configurations
	kubeConf, err := os.ReadFile(c.kubeConfigFile)
	if err != nil {
		log.Fatal(fmt.Errorf("could not load KubeConfig: %v", err))
	}

	// rook-ceph operator values
	rookCephOperatorValues := fmt.Sprintf("%s/%s", c.chartsDir, rookCephOperatorChartValues)
	valuesRookOperatorYaml, err := os.ReadFile(rookCephOperatorValues)
	if err != nil {
		log.Fatal(fmt.Errorf("could not load values.yaml for rook-operator: %v", err))
	}

	// rook-ceph cluster values
	rookCephClusterValues := fmt.Sprintf("%s/%s", c.chartsDir, rookCephClusterChartValues)
	valuesRookClusterYaml, err := os.ReadFile(rookCephClusterValues)
	if err != nil {
		log.Fatal(fmt.Errorf("could not load values.yaml for rook-cluster: %v", err))
	}

	fmt.Println("======== Preparing helm .. ===========")

	opt := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
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
		panic(err)
	}

	chartRepo := repo.Entry{
		Name: "rook-release",
		URL:  "https://charts.rook.io/release",
	}

	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		panic(err)
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName:     "rook-ceph",
		ChartName:       "rook-release/rook-ceph",
		Namespace:       "rook-ceph",
		CreateNamespace: true,
		UpgradeCRDs:     true,
		ValuesYaml:      string(valuesRookOperatorYaml),
	}

	fmt.Println("========== installing Ceph Operator Helm Chart =======")

	release, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("========= Ceph Operator Deployed =================")
	fmt.Printf("release: %+v\n", release.Name)

	//helm install --create-namespace --namespace rook-ceph rook-ceph-cluster \
	//--set operatorNamespace=rook-ceph rook-release/rook-ceph-cluster -f values.yaml
	chartClusterSpec := &helmclient.ChartSpec{
		ReleaseName:     "rook-ceph-cluster",
		ChartName:       "rook-release/rook-ceph-cluster",
		Namespace:       "rook-ceph",
		ValuesYaml:      string(valuesRookClusterYaml),
		CreateNamespace: true,

		Force: true,
	}
	fmt.Println("========== installing Ceph GetCluster Helm Chart  =======")
	clusterRelease, err := helmClient.InstallOrUpgradeChart(context.Background(), chartClusterSpec, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("========= Ceph GetCluster Deployed =================")
	fmt.Printf("release: %+v\n", clusterRelease.Name)

	releases, err := helmClient.ListDeployedReleases()
	if err != nil {
		panic(err)
	}

	fmt.Println("========= List of charts installed =================")

	for _, release := range releases {
		fmt.Printf("name: %s notes: %s\n", release.Name, release.Info.Notes)
	}
	fmt.Println("================= CHARTS installed!=========")
}
