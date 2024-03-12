package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ClusterPhase int

const (
	ClusterPhaseStarted          ClusterPhase = 0
	ClusterPhaseSetupInit        ClusterPhase = 10
	ClusterPhaseSetupDone        ClusterPhase = 15
	ClusterPhaseSshInit          ClusterPhase = 20
	ClusterPhaseSshDone          ClusterPhase = 25
	ClusterPhaseTerraformInit    ClusterPhase = 30
	ClusterPhaseTerraformDone    ClusterPhase = 35
	ClusterPhaseKubeOneInit      ClusterPhase = 40
	ClusterPhaseKubeOneDone      ClusterPhase = 45
	ClusterPhaseProviderConfInit ClusterPhase = 50
	ClusterPhaseProviderConfDone ClusterPhase = 55
	ClusterPhaseClusterReady     ClusterPhase = 100
	ClusterPhaseInstallCephInit  ClusterPhase = 110
	ClusterPhaseInstallCephDone  ClusterPhase = 120

	// Make sure we have enough room for any other non-destructive cluster phases
	ClusterPhaseDeletingCluster ClusterPhase = 900
	ClusterPhaseClusterDeleted  ClusterPhase = 999
)

type ClusterState struct {
	ID        uuid.UUID    `db:"id"`
	ClusterID uuid.UUID    `db:"cluster_id"`
	Phase     ClusterPhase `db:"phase"`
	Cluster   *Cluster
}

func NewClusterState(cluster *Cluster) *ClusterState {
	return &ClusterState{
		ID:        uuid.New(),
		ClusterID: cluster.ID,
		Phase:     ClusterPhaseStarted,
		Cluster:   cluster,
	}
}

func (cs *ClusterState) Serialize() ([]byte, error) {
	body, err := json.Marshal(cs)
	if err != nil {
		return nil, err
	}
	return body, nil
}
