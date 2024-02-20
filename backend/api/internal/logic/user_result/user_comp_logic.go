package user_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserCompLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserCompLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserCompLogic {
	return &UserCompLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserCompLogic) UserComp(req *types.UserCompReq) (resp *types.UserCompResp, err error) {
	// todo: add your logic here and delete this line

	return
}
