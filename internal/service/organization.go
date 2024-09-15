package service

import (
	"context"
	"fmt"
	"log/slog"

	"tender-service/internal/entity"
	"tender-service/internal/repo"
	"tender-service/internal/repo/repoerrs"
)

type OrganizationService struct {
	organizationRepo repo.Organization
}

func NewOrganizationService(organizationRepo repo.Organization) *OrganizationService {
	return &OrganizationService{organizationRepo: organizationRepo}
}

func (s *OrganizationService) Create(
	ctx context.Context, log *slog.Logger, input OrganizationCreateInput,
) (entity.Organization, error) {
	log.Info(fmt.Sprintf("Service - OrganizationService - Create"))
	organization := entity.Organization{
		Name:             input.Name,
		Description:      input.Description,
		OrganizationType: input.OrganizationType,
	}
	output, err := s.organizationRepo.Create(ctx, organization)
	if err != nil {
		if err == repoerrs.ErrAlreadyExists {
			return entity.Organization{}, ErrOrgAlreadyExists
		}
		log.Error(fmt.Sprintf("Service - OrganizationService - Create: %v", err))
		return entity.Organization{}, ErrCannotCreateOrg
	}
	log.Info(fmt.Sprintf("Service - OrganizationService - organizationRepo.Create - id: %s", output.Id))
	return output, nil
}

func (s *OrganizationService) Get(
	ctx context.Context, log *slog.Logger, input OrganizationGetInput,
) (entity.Organization, error) {
	output, err := s.organizationRepo.GetById(ctx, input.Id)
	if err != nil {
		if err == repoerrs.ErrNotFound {
			return entity.Organization{}, ErrOrgNotFound
		}
		log.Error(fmt.Sprintf("Service - OrganizationService - GetById: %v", err))
		return entity.Organization{}, ErrCannotGetOrg
	}
	return output, nil
}
