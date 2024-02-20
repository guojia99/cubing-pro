package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FixedNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFixedNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FixedNotifyLogic {
	return &FixedNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FixedNotifyLogic) FixedNotify(req *types.FixedNotifyReq) (resp *types.FixedNotifyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
