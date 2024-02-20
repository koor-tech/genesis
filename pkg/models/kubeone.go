package models

type Versions struct {
	Kubernetes string `yaml:"kubernetes"`
}

type CloudProviderHetzner struct {
}

type CloudProvider struct {
	Hetzner  CloudProviderHetzner `yaml:"hetzner"`
	External bool                 `yaml:"external"`
}

type KubeOneConfig struct {
	ApiVersion    string        `yaml:"apiVersion"`
	Kind          string        `yaml:"kind"`
	Versions      Versions      `yaml:"versions"`
	CloudProvider CloudProvider `yaml:"cloudProvider"`
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
