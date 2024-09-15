package service

import (
	"context"
	"fmt"
	"log/slog"

	"tender-service/internal/entity"
	"tender-service/internal/repo"
	"tender-service/internal/repo/repoerrs"
)

type BidService struct {
	bidRepo repo.Bid
}

func NewBidService(bidRepo repo.Bid) *BidService {
	return &BidService{bidRepo: bidRepo}
}

func (s *BidService) Create(
	ctx context.Context, log *slog.Logger, input BidCreateInput,
) (entity.Bid, error) {
	log.Info(fmt.Sprintf("Service - BidService - Create"))
	bid := entity.Bid{
		Name:        input.Name,
		Description: input.Description,
		TenderId:    input.TenderId,
		AuthorType:  input.AuthorType,
		AuthorId:    input.AuthorId,
	}
	output, err := s.bidRepo.Create(ctx, bid)
	if err != nil {
		if err == repoerrs.ErrAlreadyExists {
			return entity.Bid{}, ErrBidAlreadyExists
		}
		log.Error(fmt.Sprintf("Service - BidService - Create: %v", err))
		return entity.Bid{}, ErrCannotGetBid
	}
	log.Info(fmt.Sprintf("Service - BidService - Create - id: %s", output.Id))
	return output, nil
}

func (s *BidService) GetByTenderId(
	ctx context.Context, log *slog.Logger, input BidGetByTenderIdInput,
) ([]entity.Bid, error) {
	output, err := s.bidRepo.GetByTenderID(ctx, input.Limit, input.Offset, input.UserId, input.TenderId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - BidService - GetByTenderId: %v", err))
		return nil, ErrCannotGetBid
	}
	return output, nil
}

func (s *BidService) GetById(
	ctx context.Context, log *slog.Logger, bidId string,
) (entity.Bid, error) {
	output, err := s.bidRepo.GetById(ctx, bidId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - BidService - GetById: %v", err))
		return entity.Bid{}, ErrCannotGetBid
	}
	return output, nil
}

func (s *BidService) GetMy(
	ctx context.Context, log *slog.Logger, input BidGetMyInput,
) ([]entity.Bid, error) {
	//log.Info(fmt.Sprintf("limit - %d offset - %d", input.Limit, input.Offset))
	output, err := s.bidRepo.GetMyPagination(ctx, input.Limit, input.Offset, input.UserId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - BidService - GetMy: %v", err))
		return nil, ErrCannotGetBid
	}
	return output, nil
}

func (s *BidService) PutStatus(ctx context.Context, log *slog.Logger, bidId, status string) (entity.Bid, error) {
	err := s.bidRepo.PutStatus(ctx, bidId, status)
	if err != nil {
		log.Error(fmt.Sprintf("Service - BidService - PutStatus: %v", err))
		return entity.Bid{}, ErrCannotPutStatus
	}

	err = s.bidRepo.IncrementVersion(ctx, bidId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - BidService - IncrementVersion: %v", err))
		return entity.Bid{}, ErrCannotIncrement
	}

	output, err := s.GetById(ctx, log, bidId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - BidService - GetById: %v", err))
		return entity.Bid{}, ErrCannotGetBid
	}
	return output, nil
}

func (s *BidService) EditBid(ctx context.Context, log *slog.Logger, input BidEditInput, bidId string) (
	entity.Bid, error,
) {
	log.Info("EditBid")
	var err error
	in := entity.Bid{
		Name:        input.Name,
		Description: input.Description,
	}

	if err = s.bidRepo.EditBid(ctx, in, bidId); err != nil {
		log.Error(fmt.Sprintf("Service - BidService - EditBid: %v", err))
		return entity.Bid{}, ErrCannotEditBid
	}

	err = s.bidRepo.IncrementVersion(ctx, bidId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - BidService - IncrementVersion: %v", err))
		return entity.Bid{}, ErrCannotIncrement
	}

	outputNew, err := s.GetById(ctx, log, bidId)
	if err != nil {
		log.Error(fmt.Sprintf("Service - BidService - GetById: %v", err))
		return entity.Bid{}, ErrCannotGetBid
	}
	return outputNew, nil
}
