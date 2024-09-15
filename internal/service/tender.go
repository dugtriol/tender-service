package service

import (
	"context"
	"fmt"
	"log/slog"

	"tender-service/internal/entity"
	"tender-service/internal/repo"
	"tender-service/internal/repo/repoerrs"
)

type TenderService struct {
	tenderRepo repo.Tender
}

func NewTenderService(tenderRepo repo.Tender) *TenderService {
	return &TenderService{tenderRepo: tenderRepo}
}

func (s *TenderService) Create(
	ctx context.Context, log *slog.Logger, input TenderCreateInput,
) (entity.Tender, error) {
	log.Info(fmt.Sprintf("Service - TenderService - Create"))
	tender := entity.Tender{
		Name:            input.Name,
		Description:     input.Description,
		ServiceType:     input.ServiceType,
		OrganizationId:  input.OrganizationId,
		CreatorUsername: input.CreatorUsername,
	}
	output, err := s.tenderRepo.Create(ctx, tender)
	if err != nil {
		if err == repoerrs.ErrAlreadyExists {
			return entity.Tender{}, ErrTenderAlreadyExists
		}
		log.Error(fmt.Sprintf("Service - TenderService - Create: %v", err))
		return entity.Tender{}, ErrCannotGetTender
	}
	log.Info(fmt.Sprintf("Service - TenderService - tenderRepo.Create - id: %s", output.Id))
	return output, nil
}

func (s *TenderService) GetByType(
	ctx context.Context, log *slog.Logger, input TenderGetByTypeInput,
) ([]entity.Tender, error) {
	output, err := s.tenderRepo.GetByTypePagination(ctx, input.Limit, input.Offset, input.ServiceType)
	if err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - GetByTenderId: %v", err))
		return nil, ErrCannotGetTender
	}
	return output, nil
}

func (s *TenderService) GetMy(
	ctx context.Context, log *slog.Logger, input TenderGetMyInput,
) ([]entity.Tender, error) {
	log.Info(fmt.Sprintf("limit - %d offset - %d", input.Limit, input.Offset))
	output, err := s.tenderRepo.GetMyPagination(ctx, input.Limit, input.Offset, input.Username)
	if err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - GetMy: %v", err))
		return nil, ErrCannotGetTender
	}
	return output, nil
}

func (s *TenderService) GetById(
	ctx context.Context, log *slog.Logger, id string,
) (entity.Tender, error) {
	// TODO log
	log.Info("GetById")
	output, err := s.tenderRepo.GetById(ctx, id)
	if err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - GetMy: %v", err))
		return entity.Tender{}, ErrCannotGetTender
	}
	return output, nil
}

func (s *TenderService) PutStatus(ctx context.Context, log *slog.Logger, tenderId, status string) (
	entity.Tender, error,
) {
	err := s.tenderRepo.PutStatus(ctx, tenderId, status)
	if err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - PutStatus: %v", err))
		return entity.Tender{}, ErrCannotPutStatus
	}

	err = s.tenderRepo.IncrementVersion(ctx, tenderId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - IncrementVersion: %v", err))
		return entity.Tender{}, ErrCannotIncrement
	}

	output, err := s.GetById(ctx, log, tenderId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - GetById: %v", err))
		return entity.Tender{}, ErrCannotGetTender
	}
	return output, nil
}

func (s *TenderService) EditTender(
	ctx context.Context, log *slog.Logger, input TenderEditInput, tenderId string,
) (entity.Tender, error) {
	var err error
	in := entity.Tender{
		Name:        input.Name,
		Description: input.Description,
		ServiceType: input.ServiceType,
	}

	if err = s.tenderRepo.EditTender(ctx, in, tenderId); err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - EditBid: %v", err))
		return entity.Tender{}, ErrCannotEditTender
	}

	err = s.tenderRepo.IncrementVersion(ctx, tenderId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - IncrementVersion: %v", err))
		return entity.Tender{}, ErrCannotIncrement
	}

	outputNew, err := s.GetById(ctx, log, tenderId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - TenderService - EditBid: %v", err))
		return entity.Tender{}, ErrCannotGetTender
	}
	return outputNew, nil
}
