package biz

import (
	"context"

	pb "moredoc/api/v1"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/logger"
	"moredoc/util"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type NavigationAPIService struct {
	pb.UnimplementedNavigationAPIServer
	dbModel *model.DBModel
	logger  logger.Logger
}

func NewNavigationAPIService(dbModel *model.DBModel, logger logger.Logger) (service *NavigationAPIService) {
	return &NavigationAPIService{dbModel: dbModel, logger: logger}
}

func (s *NavigationAPIService) checkPermission(ctx context.Context) (userClaims *auth.UserClaims, err error) {
	return checkGRPCPermission(s.dbModel, ctx)
}

func (s *NavigationAPIService) CreateNavigation(ctx context.Context, req *pb.Navigation) (*pb.Navigation, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	nav := &model.Navigation{}
	err = util.CopyStruct(req, nav)
	if err != nil {
		s.logger.Errorf("CopyStruct", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	s.logger.Debugf("CreateNavigation", zap.Any("nav", nav), zap.Any("req", req))

	err = s.dbModel.CreateNavigation(nav)
	if err != nil {
		s.logger.Errorf("CreateNavigation", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res := &pb.Navigation{}
	err = util.CopyStruct(nav, res)
	if err != nil {
		s.logger.Errorf("CopyStruct", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return res, nil
}

func (s *NavigationAPIService) UpdateNavigation(ctx context.Context, req *pb.Navigation) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	nav := &model.Navigation{}
	err = util.CopyStruct(req, nav)
	if err != nil {
		s.logger.Errorf("CopyStruct", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	s.logger.Debugf("UpdateNavigation", zap.Any("nav", nav), zap.Any("req", req))

	err = s.dbModel.UpdateNavigation(nav)
	if err != nil {
		s.logger.Errorf("UpdateNavigation", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *NavigationAPIService) DeleteNavigation(ctx context.Context, req *pb.DeleteNavigationRequest) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	err = s.dbModel.DeleteNavigation(req.Id)
	if err != nil {
		s.logger.Errorf("DeleteNavigation", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *NavigationAPIService) GetNavigation(ctx context.Context, req *pb.GetNavigationRequest) (*pb.Navigation, error) {
	nav := &model.Navigation{}
	err := s.dbModel.GetByID(req.Id, nav)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Errorf("GetNavigation", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res := &pb.Navigation{}
	err = util.CopyStruct(nav, res)
	if err != nil {
		s.logger.Errorf("CopyStruct", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return res, nil
}
