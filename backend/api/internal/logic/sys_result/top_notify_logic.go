package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TopNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTopNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TopNotifyLogic {
	return &TopNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TopNotifyLogic) TopNotify(req *types.TopNotifyReq) (resp *types.TopNotifyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
