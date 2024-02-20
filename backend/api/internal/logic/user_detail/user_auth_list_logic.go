package user_detail

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserAuthListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserAuthListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserAuthListLogic {
	return &UserAuthListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserAuthListLogic) UserAuthList(req *types.UserAuthListReq) (resp *types.UserAuthListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
