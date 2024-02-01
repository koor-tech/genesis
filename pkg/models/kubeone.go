package models

type Versions struct {
	Kubernetes string `json:"kubernetes"`
}

type CloudProviderHetzner struct {
}

type CloudProvider struct {
	Hetzner  CloudProviderHetzner `json:"hetzner"`
	External bool                 `json:"external"`
}

type KubeOneConfig struct {
	ApiVersion    string        `json:"apiVersion"`
	Kind          string        `json:"kind"`
	Versions      Versions      `json:"versions"`
	CloudProvider CloudProvider `json:"cloud_provider"`
}

func NewKubeOneConfig() *KubeOneConfig {
	return &KubeOneConfig{
		ApiVersion: "kubeone.k8c.io/v1beta2",
		Kind:       "KubeOneCluster",
		Versions: Versions{
			Kubernetes: "1.25.6",
		},
		CloudProvider: CloudProvider{
			Hetzner:  CloudProviderHetzner{},
			External: true,
		},
	}
}
