package models

type TerraformConfig struct {
	ClusterName                      string `json:"cluster_name"`
	SshPublicKeyFile                 string `json:"ssh_public_key_file"`
	ControlPlaneVmCount              int    `json:"control_plane_vm_count"`
	InitialMachineDeploymentReplicas int    `json:"initial_machinedeployment_replicas"`
	WorkerType                       string `json:"worker_type"`
	ControlPlaneType                 string `json:"control_plane_type"`
	Os                               string `json:"os"`
	WorkerOs                         string `json:"worker_os"`
}

func NewTerraformConfig(cluster *Cluster, dst string) *TerraformConfig {
	return &TerraformConfig{
		ClusterName:                      "koor-client-" + cluster.Customer.Company,
		SshPublicKeyFile:                 dst + "/id_ed25519.pub",
		ControlPlaneVmCount:              1,
		InitialMachineDeploymentReplicas: 4,
		WorkerType:                       "cpx41",
		ControlPlaneType:                 "cpx31",
		Os:                               "ubuntu",
		WorkerOs:                         "ubuntu",
	}
}
