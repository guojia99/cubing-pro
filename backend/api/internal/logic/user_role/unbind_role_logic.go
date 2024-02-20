package user_role

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnbindRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnbindRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbindRoleLogic {
	return &UnbindRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnbindRoleLogic) UnbindRole(req *types.UnbindRoleReq) (resp *types.UnbindRoleResp, err error) {
	// todo: add your logic here and delete this line

	return
}
