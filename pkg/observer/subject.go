package observer

import "github.com/koor-tech/genesis/pkg/models"

type Subject struct {
	observers []func(state *models.ClusterState) error
}

func NewSubject() *Subject {
	return &Subject{}
}

// AddObserver adds a new observer function
func (s *Subject) AddObserver(observer func(*models.ClusterState) error) {
	s.observers = append(s.observers, observer)
}

// Notify all observers
func (s *Subject) Notify(state *models.ClusterState) error {
	for _, observer := range s.observers {
		err := observer(state)
		if err != nil {
			return err
		}
	}
	return nil
}
