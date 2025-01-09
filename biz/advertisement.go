package biz

import (
	"context"

	pb "moredoc/api/v1"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/pkg/logger"
	"moredoc/util"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type AdvertisementAPIService struct {
	pb.UnimplementedAdvertisementAPIServer
	dbModel *model.DBModel
	logger  logger.Logger
}

func NewAdvertisementAPIService(dbModel *model.DBModel, logger logger.Logger) (service *AdvertisementAPIService) {
	return &AdvertisementAPIService{dbModel: dbModel, logger: logger}
}

func (s *AdvertisementAPIService) checkPermission(ctx context.Context) (userClaims *auth.UserClaims, err error) {
	return checkGRPCPermission(s.dbModel, ctx)
}

func (s *AdvertisementAPIService) CreateAdvertisement(ctx context.Context, req *pb.Advertisement) (*pb.Advertisement, error) {
	userCliams, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	adv := &model.Advertisement{}
	if err = util.CopyStruct(req, adv); err != nil {
		s.logger.Errorf("CreateAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "创建广告失败:"+err.Error())
	}

	if adv.Position == "" || adv.Content == "" {
		return nil, status.Errorf(codes.InvalidArgument, "广告位和广告内容均不能为空")
	}

	adv.UserId = userCliams.UserId
	if err = s.dbModel.Create(adv); err != nil {
		s.logger.Errorf("CreateAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "创建广告失败:"+err.Error())
	}

	res := &pb.Advertisement{}
	if err = util.CopyStruct(adv, res); err != nil {
		s.logger.Errorf("CreateAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return res, nil
}

func (s *AdvertisementAPIService) UpdateAdvertisement(ctx context.Context, req *pb.Advertisement) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	adv := &model.Advertisement{}
	if err = util.CopyStruct(req, adv); err != nil {
		s.logger.Errorf("UpdateAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "更新广告失败:"+err.Error())
	}

	err = s.dbModel.UpdateByFields(adv, model.TableAdvertisement, adv.Id)
	if err != nil {
		s.logger.Errorf("UpdateAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "更新广告失败:"+err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *AdvertisementAPIService) DeleteAdvertisement(ctx context.Context,
	req *pb.DeleteAdvertisementRequest) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.Id) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "广告id不能为空")
	}

	err = s.dbModel.DeleteByIds(req.Id, &model.Advertisement{})
	if err != nil {
		s.logger.Errorf("DeleteAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "删除广告失败:"+err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *AdvertisementAPIService) GetAdvertisement(ctx context.Context, req *pb.GetAdvertisementRequest) (*pb.Advertisement, error) {
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "广告id不能为空")
	}

	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	ad := &model.Advertisement{}
	err = s.dbModel.GetByID(req.Id, ad)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Errorf("GetAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "获取广告失败:"+err.Error())
	}

	res := &pb.Advertisement{}
	err = util.CopyStruct(ad, res)
	if err != nil {
		s.logger.Errorf("GetAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return res, nil
}

func (s *AdvertisementAPIService) ListAdvertisement(ctx context.Context, req *pb.ListAdvertisementRequest) (*pb.ListAdvertisementReply, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	opt := &model.OptionGetList{
		WithCount: true,
		Page:      int(req.Page),
		Size:      int(req.Size_),
		QueryIn:   make(map[string][]interface{}),
		QueryLike: make(map[string][]interface{}),
	}

	if len(req.Position) > 0 {
		opt.QueryIn["position"] = cvt.ToArray(req.Position)
	}

	if len(req.Enable) > 0 {
		opt.QueryIn["enable"] = cvt.ToArray(req.Enable)
	}

	if req.Wd != "" {
		fields := []string{"title", "content", "remark"}
		for _, field := range fields {
			opt.QueryLike[field] = []interface{}{req.Wd}
		}
	}

	advs, total, err := s.dbModel.GetAdvertisementList(opt)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Errorf("ListAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "获取广告列表失败:"+err.Error())
	}

	res := &pb.ListAdvertisementReply{
		Total: total,
	}
	err = util.CopyStruct(advs, &res.Advertisement)
	if err != nil {
		s.logger.Errorf("ListAdvertisement", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return res, nil
}
