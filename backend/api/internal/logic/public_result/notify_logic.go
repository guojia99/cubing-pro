package public_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type NotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotifyLogic {
	return &NotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NotifyLogic) Notify(req *types.NotifyReq) (resp *types.NotifyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
