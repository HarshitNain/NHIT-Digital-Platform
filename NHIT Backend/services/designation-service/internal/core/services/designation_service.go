package services

import (
	"context"
	"log/slog"

	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/ports"
	"github.com/google/uuid"
)

type designationService struct {
	repo   ports.DesignationRepository
	logger *slog.Logger
}

func NewDesignationService(repo ports.DesignationRepository, logger *slog.Logger) ports.DesignationService {
	return &designationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *designationService) CreateDesignation(ctx context.Context, name, description string, orgID *uuid.UUID) (*domain.Designation, error) {
	// Check if already exists in this org
	exists, err := s.repo.Exists(ctx, name, orgID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrDesignationAlreadyExists // Ensure this exists in domain
	}

	// domain.NewDesignation sets ID, CreatedAt, UpdatedAt
	d, err := domain.NewDesignation(name, description, orgID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *designationService) GetDesignation(ctx context.Context, id uuid.UUID) (*domain.Designation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *designationService) UpdateDesignation(ctx context.Context, id uuid.UUID, name, description string) (*domain.Designation, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := d.Update(name, description); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *designationService) DeleteDesignation(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *designationService) ListDesignations(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Designation, int64, error) {
	list, total, err := s.repo.List(ctx, orgID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
