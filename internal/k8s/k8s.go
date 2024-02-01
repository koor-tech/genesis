package k8s

import (
	"context"
	"fmt"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"os"
)

type Cluster struct {
	kubeConfigFile string
}

func New(kubeConfigFile string) *Cluster {
	return &Cluster{
		kubeConfigFile: kubeConfigFile,
	}
	//// get kubernetes client
	//client, _ := kubernetes.NewForConfig(config)
	//client = client
}

func (c *Cluster) InstallCharts() {
	fmt.Println("======== Installing helm charts running... ===========")
	// Read the kubeconfig file
	kubeConf, err := os.ReadFile(c.kubeConfigFile)
	if err != nil {
		log.Fatal(fmt.Errorf("could not load KubeConfig: %v", err))
	}

	fmt.Println("======== Preparing helm .. ===========")

	opt := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			//Namespace:        "rook-ceph",
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

	fmt.Println("==========================")
	fmt.Printf("helmClient: %+v\n", helmClient)
	fmt.Println("==========================")
	fmt.Println("adding chart")
	chartRepo := repo.Entry{
		Name: "rook-release",
		URL:  "https://charts.rook.io/release",
	}

	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		panic(err)
	}

	fmt.Println("==========================")
	fmt.Println("helm chart rook-ceph added")

	valuesRookOperatorYaml, err := os.ReadFile("/home/javier/src/koor-platform/charts/rook-ceph/values.yaml")
	if err != nil {
		log.Fatal(fmt.Errorf("could not load KubeConfig: %v", err))
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

	//fmt.Println("========= removing Ceph Operator =================")
	//err = helmClient.RollbackRelease(chartSpec)
	//if err != nil {
	//	panic(err)
	//}

	valuesRookClusterYaml, err := os.ReadFile("/home/javier/src/koor-platform/charts/rook-ceph-cluster/values.yaml")
	if err != nil {
		log.Fatal(fmt.Errorf("could not load KubeConfig: %v", err))
	}
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
	fmt.Println("========== installing Ceph Cluster Helm Chart  =======")
	clusterRelease, err := helmClient.InstallOrUpgradeChart(context.Background(), chartClusterSpec, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("========= Ceph Cluster Deployed =================")
	fmt.Printf("release: %+v\n", clusterRelease.Name)
	//fmt.Println("========= removing Ceph Cluster =================")
	//err = helmClient.RollbackRelease(chartClusterSpec)
	//if err != nil {
	//	panic(err)
	//}

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
