package user_role

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleBindUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleBindUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleBindUsersLogic {
	return &RoleBindUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleBindUsersLogic) RoleBindUsers(req *types.RoleBindUsersReq) (resp *types.RoleBindUsersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
