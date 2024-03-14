package observer

import (
	"sync"

	"github.com/koor-tech/genesis/pkg/models"
)

type ObserverFn = func(*models.ClusterState) error

type Subject struct {
	mu sync.Mutex

	observers []ObserverFn
}

func NewSubject() *Subject {
	return &Subject{
		mu: sync.Mutex{},
	}
}

// AddObserver adds a new observer function
func (s *Subject) AddObserver(observer ObserverFn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.observers = append(s.observers, observer)
}

// Notify all observers
func (s *Subject) Notify(state *models.ClusterState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, observer := range s.observers {
		err := observer(state)
		if err != nil {
			return err
		}
	}
	return nil
}
