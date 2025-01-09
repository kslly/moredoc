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
)

type FriendlinkAPIService struct {
	pb.UnimplementedFriendlinkAPIServer
	dbModel *model.DBModel
	logger  logger.Logger
}

func NewFriendlinkAPIService(dbModel *model.DBModel, logger logger.Logger) (service *FriendlinkAPIService) {
	return &FriendlinkAPIService{dbModel: dbModel, logger: logger}
}

// checkPermission 检查用户权限
func (s *FriendlinkAPIService) checkPermission(ctx context.Context) (userClaims *auth.UserClaims, err error) {
	return checkGRPCPermission(s.dbModel, ctx)
}

// CreateFriendlink 创建友情链接，需要鉴权
func (s *FriendlinkAPIService) CreateFriendlink(ctx context.Context, req *pb.Friendlink) (*pb.Friendlink, error) {
	s.logger.Debugf("CreateFriendlink", zap.Any("req", req))
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	friendlink := &model.Friendlink{}
	util.CopyStruct(req, friendlink)
	err = s.dbModel.Create(friendlink)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	pbFriendlink := &pb.Friendlink{}
	util.CopyStruct(friendlink, pbFriendlink)
	return pbFriendlink, nil
}

// UpdateFriendlink 更新友情链接，需要鉴权
func (s *FriendlinkAPIService) UpdateFriendlink(ctx context.Context, req *pb.Friendlink) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "参数不正确")
	}

	friendlink := &model.Friendlink{}
	util.CopyStruct(req, friendlink)
	err = s.dbModel.UpdateByFields(friendlink, model.TableFriendlink, int64(friendlink.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// DeleteFriendlink 删除友情链接，需要鉴权
func (s *FriendlinkAPIService) DeleteFriendlink(ctx context.Context, req *pb.DeleteFriendlinkRequest) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	err = s.dbModel.DeleteByIds(req.Id, &model.Friendlink{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// GetFriendlink 查询友情链接
func (s *FriendlinkAPIService) GetFriendlink(ctx context.Context, req *pb.GetFriendlinkRequest) (*pb.Friendlink, error) {
	// _, err := s.checkPermission(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	friendlink := &model.Friendlink{}
	err := s.dbModel.GetByID(req.Id, friendlink)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	pbFriendlink := &pb.Friendlink{}
	util.CopyStruct(friendlink, pbFriendlink)

	s.logger.Debugf("GetFriendlink", zap.Any("pbFriendlink", pbFriendlink), zap.Any("friendlink", friendlink))

	return pbFriendlink, nil
}
