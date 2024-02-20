package public_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type EventsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EventsLogic {
	return &EventsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EventsLogic) Events(req *types.EventsReq) (resp *types.EventsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
