package user_role

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindRoleLogic {
	return &BindRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindRoleLogic) BindRole(req *types.BindRoleReq) (resp *types.BindRoleResp, err error) {
	// todo: add your logic here and delete this line

	return
}
