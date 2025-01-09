package biz

import (
	"context"

	pb "moredoc/api/v1"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/logger"
	"moredoc/util"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BannerAPIService struct {
	pb.UnimplementedBannerAPIServer
	dbModel *model.DBModel
	logger  logger.Logger
}

func NewBannerAPIService(dbModel *model.DBModel, logger logger.Logger) (service *BannerAPIService) {
	return &BannerAPIService{dbModel: dbModel, logger: logger}
}

func (s *BannerAPIService) checkPermission(ctx context.Context) (*auth.UserClaims, error) {
	return checkGRPCPermission(s.dbModel, ctx)
}

// CreateBanner 创建轮播图
func (s *BannerAPIService) CreateBanner(ctx context.Context, req *pb.Banner) (*pb.Banner, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	var banner model.Banner
	util.CopyStruct(req, &banner)
	err = s.dbModel.Create(&banner)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbBanner := &pb.Banner{}
	util.CopyStruct(&banner, pbBanner)

	return pbBanner, nil
}

// UpdateBanner 更新轮播图
func (s *BannerAPIService) UpdateBanner(ctx context.Context, req *pb.Banner) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	var banner model.Banner
	util.CopyStruct(req, &banner)
	err = s.dbModel.UpdateByFields(&banner, model.TableBanner, banner.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *BannerAPIService) DeleteBanner(ctx context.Context, req *pb.DeleteBannerRequest) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	err = s.dbModel.DeleteBanner(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *BannerAPIService) GetBanner(ctx context.Context, req *pb.GetBannerRequest) (*pb.Banner, error) {
	if _, errPermission := s.checkPermission(ctx); errPermission != nil {
		return nil, errPermission
	}

	banner := &model.Banner{}
	err := s.dbModel.GetByID(req.Id, banner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pbBanner := &pb.Banner{}
	util.CopyStruct(banner, pbBanner)
	return pbBanner, nil
}
