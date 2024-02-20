package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateEventLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateEventLogic {
	return &CreateEventLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateEventLogic) CreateEvent(req *types.CreateEventReq) (resp *types.CreateEventResp, err error) {
	// todo: add your logic here and delete this line

	return
}
