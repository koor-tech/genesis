package cluster

import (
	"context"
	"github.com/koor-tech/genesis/pkg/models"
)

func (k *KoorCluster) Listen(ctx context.Context) {
	k.subject.AddObserver(func(clusterState *models.ClusterState) error {
		k.logger.Info("Running observer data will be send to rabbitmq")
		serializedState, err := clusterState.Serialize()
		if err != nil {
			k.logger.Error("unable to serialize cluster state", "err", err)
			return err
		}
		k.queue.Publish(ctx, serializedState)
		return nil
	})
}
