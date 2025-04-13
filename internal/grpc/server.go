package grpc

import (
	"context"

	"github.com/devWaylander/pvz_store/api"
	"github.com/devWaylander/pvz_store/internal/pb/pvz_v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Repository interface {
	GetPVZs(ctx context.Context) ([]api.PVZ, error)
}

type grpcService struct {
	pvz_v1.UnimplementedPVZServiceServer
	repo Repository
}

func New(repo Repository) *grpcService {
	return &grpcService{repo: repo}
}

func (s *grpcService) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	pvzs, err := s.repo.GetPVZs(ctx)
	if err != nil {
		return nil, err
	}

	var response pvz_v1.GetPVZListResponse
	for _, pvz := range pvzs {
		response.Pvzs = append(response.Pvzs, &pvz_v1.PVZ{
			Id:               pvz.Id.String(),
			City:             string(pvz.City),
			RegistrationDate: timestamppb.New(*pvz.RegistrationDate),
		})
	}

	return &response, nil
}
