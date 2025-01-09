package biz

import (
	"context"
	"net/http"
	"strings"

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
)

type ginResponse struct {
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Code    int         `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type AttachmentAPIService struct {
	pb.UnimplementedAttachmentAPIServer
	dbModel *model.DBModel
	logger  logger.Logger
}

var errorHash = map[string]interface{}{
	"code":    http.StatusBadRequest,
	"message": "hash值必须32位",
}

func NewAttachmentAPIService(dbModel *model.DBModel, logger logger.Logger) (service *AttachmentAPIService) {
	return &AttachmentAPIService{dbModel: dbModel, logger: logger}
}

// checkPermission 检查用户权限
func (s *AttachmentAPIService) checkPermission(ctx context.Context) (userClaims *auth.UserClaims, err error) {
	return checkGRPCPermission(s.dbModel, ctx)
}

// UpdateAttachment 更新附件。只允许更新附件名称、是否合法以及描述字段
func (s *AttachmentAPIService) UpdateAttachment(ctx context.Context, req *pb.Attachment) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	err = s.dbModel.UpdateByFields(&model.Attachment{
		Id:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Enable:      req.Enable,
	}, model.TableAttachment, req.Id, "name", "enable", "description")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *AttachmentAPIService) DeleteAttachment(ctx context.Context, req *pb.DeleteAttachmentRequest) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	err = s.dbModel.DeleteByIds(req.Id, &model.Attachment{})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// GetAttachment 查询单个附件信息
func (s *AttachmentAPIService) GetAttachment(ctx context.Context, req *pb.GetAttachmentRequest) (*pb.Attachment, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	attachment, err := s.dbModel.GetAttachment(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbAttachment := &pb.Attachment{}
	util.CopyStruct(&attachment, pbAttachment)

	return pbAttachment, nil
}

func (s *AttachmentAPIService) ListAttachment(ctx context.Context, req *pb.ListAttachmentRequest) (*pb.ListAttachmentReply, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	opt := &model.OptionGetList{
		Page:      int(req.Page),
		Size:      int(req.Size_),
		WithCount: true,
		QueryIn:   make(map[string][]interface{}),
	}

	if len(req.UserId) > 0 {
		opt.QueryIn["user_id"] = cvt.ToArray(req.UserId)
	}

	if len(req.Enable) > 0 {
		opt.QueryIn["enable"] = cvt.ToArray(req.Enable)
	}

	if len(req.Type) > 0 {
		opt.QueryIn["type"] = cvt.ToArray(req.Type)
	}

	req.Wd = strings.TrimSpace(req.Wd)
	if req.Wd != "" {
		wd := "%" + req.Wd + "%"
		opt.QueryLike = map[string][]interface{}{"name": {wd}, "description": {wd}}
	}

	attachments, total, err := s.dbModel.GetAttachmentList(opt)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var pbAttachments []*pb.Attachment
	util.CopyStruct(&attachments, &pbAttachments)

	var (
		userIds        []int64
		userIdIndexMap = make(map[int64][]int)
	)

	for idx, attchment := range pbAttachments {
		attchment.TypeName = s.dbModel.GetAttachmentTypeName(int(attchment.Type))
		userIds = append(userIds, attchment.UserId)
		userIdIndexMap[attchment.UserId] = append(userIdIndexMap[attchment.UserId], idx)
		pbAttachments[idx] = attchment
	}

	if size := len(userIds); size > 0 {
		users, _, _ := s.dbModel.GetUserList(&model.OptionGetList{Ids: userIds, Page: 1, Size: size, SelectFields: []string{"id", "username"}})
		s.logger.Debugf("GetUserList", zap.Any("users", users))
		for _, user := range users {
			if indexes, ok := userIdIndexMap[user.Id]; ok {
				for _, idx := range indexes {
					pbAttachments[idx].Username = user.Username
				}
			}
		}
	}
	return &pb.ListAttachmentReply{Total: total, Attachment: pbAttachments}, nil
}
