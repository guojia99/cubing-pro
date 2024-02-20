package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNotifyLogic {
	return &UpdateNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNotifyLogic) UpdateNotify(req *types.UpdateNotifyReq) (resp *types.UpdateNotifyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
