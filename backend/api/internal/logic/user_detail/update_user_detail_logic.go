package user_detail

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserDetailLogic {
	return &UpdateUserDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserDetailLogic) UpdateUserDetail(req *types.UpdateUserDetailReq) (resp *types.UpdateUserDetailResp, err error) {
	// todo: add your logic here and delete this line

	return
}
