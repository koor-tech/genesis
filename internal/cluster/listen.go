package cluster

import (
	"context"

	"github.com/koor-tech/genesis/pkg/models"
)

func (s *Service) Listen(ctx context.Context) {
	s.subject.AddObserver(func(clusterState *models.ClusterState) error {
		s.logger.Info("Running observer data will be send to rabbitmq")

		serializedState, err := clusterState.Serialize()
		if err != nil {
			s.logger.Error("unable to serialize cluster state", "err", err)
			return err
		}

		return s.queue.Publish(ctx, serializedState)
	})
}
