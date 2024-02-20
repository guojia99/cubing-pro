package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateEventLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateEventLogic {
	return &UpdateEventLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateEventLogic) UpdateEvent(req *types.UpdateEventReq) (resp *types.UpdateEventResp, err error) {
	// todo: add your logic here and delete this line

	return
}
