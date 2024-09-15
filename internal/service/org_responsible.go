package service

import (
	"context"
	"fmt"
	"log/slog"

	"tender-service/internal/entity"
	"tender-service/internal/repo"
	"tender-service/internal/repo/repoerrs"
)

type OrgResponsibleService struct {
	orgRespRepo repo.OrgResponsible
}

func NewOrgResponsibleService(orgRespRepo repo.OrgResponsible) *OrgResponsibleService {
	return &OrgResponsibleService{orgRespRepo: orgRespRepo}
}

func (s *OrgResponsibleService) Create(
	ctx context.Context, log *slog.Logger, input OrgResponsibleCreateInput,
) (entity.OrgResponsible, error) {
	log.Info(fmt.Sprintf("Service - OrganizationService - Create"))
	orgresp := entity.OrgResponsible{
		OrganizationId: input.OrganizationId,
		UserId:         input.UserId,
	}
	output, err := s.orgRespRepo.Create(ctx, orgresp)
	if err != nil {
		if err == repoerrs.ErrAlreadyExists {
			return entity.OrgResponsible{}, ErrOrgRespAlreadyExists
		}
		log.Error(fmt.Sprintf("Service - OrgResponsibleService - Create: %v", err))
		return entity.OrgResponsible{}, ErrCannotCreateOrgResp
	}
	log.Info(fmt.Sprintf("Service - OrgResponsibleService - orgRespRepo.Create - id: %s", output.Id))
	return output, nil
}

func (s *OrgResponsibleService) Get(
	ctx context.Context, log *slog.Logger, input OrgResponsibleGetInput,
) (entity.OrgResponsible, error) {
	output, err := s.orgRespRepo.GetById(ctx, input.Id)
	if err != nil {
		if err == repoerrs.ErrNotFound {
			return entity.OrgResponsible{}, ErrOrgRespNotFound
		}
		log.Error(fmt.Sprintf("Service - OrgResponsibleService - GetById: %v", err))
		return entity.OrgResponsible{}, ErrCannotGetOrgResp
	}
	return output, nil
}

func (s *OrgResponsibleService) GetByIds(
	ctx context.Context, log *slog.Logger, input OrgResponsibleGetByIdsInput,
) (entity.OrgResponsible, error) {
	ogrresp := entity.OrgResponsible{
		OrganizationId: input.OrganizationId,
		UserId:         input.UserId,
	}
	output, err := s.orgRespRepo.GetByIds(ctx, ogrresp)
	if err != nil {
		if err == repoerrs.ErrNotFound {
			return entity.OrgResponsible{}, ErrOrgRespNotFound
		}
		log.Error(fmt.Sprintf("Service - OrgResponsibleService - GetByIds: %v", err))
		return entity.OrgResponsible{}, ErrCannotGetOrgResp
	}
	return output, nil
}
