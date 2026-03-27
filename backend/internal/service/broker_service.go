package service

import (
	"trackly-backend/internal/model"
	"trackly-backend/internal/repository"
	"trackly-backend/pkg/filter"
)

type BrokerService interface {
	GetAllBrokers(opt filter.Options, allowedFields []string) ([]model.Broker, int, error)
	AddBroker(symbol string, name string) error
}

type brokerService struct {
	repository repository.BrokerRepository
}

func NewBrokerService(repository repository.BrokerRepository) BrokerService {
	return &brokerService{repository: repository}
}

func (s *brokerService) GetAllBrokers(opt filter.Options, allowedFields []string) ([]model.Broker, int, error) {
	return s.repository.FindAll(opt, allowedFields)
}

func (s *brokerService) AddBroker(symbol string, name string) error {
	return s.repository.Create(symbol, name)
}
