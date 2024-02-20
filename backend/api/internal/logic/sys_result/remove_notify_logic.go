package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveNotifyLogic {
	return &RemoveNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveNotifyLogic) RemoveNotify(req *types.RemoveNotifyReq) (resp *types.RemoveNotifyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
