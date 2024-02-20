package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveEventLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveEventLogic {
	return &RemoveEventLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveEventLogic) RemoveEvent(req *types.RemoveEventReq) (resp *types.RemoveEventResp, err error) {
	// todo: add your logic here and delete this line

	return
}
