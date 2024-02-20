package user_detail

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRetrievePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRetrievePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRetrievePasswordLogic {
	return &UserRetrievePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRetrievePasswordLogic) UserRetrievePassword(req *types.UserRetrievePasswordReq) (resp *types.UserRetrievePasswordResp, err error) {
	// todo: add your logic here and delete this line

	return
}
