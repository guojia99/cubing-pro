package user_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreResultLogic {
	return &PreResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreResultLogic) PreResult(req *types.PreResultReq) (resp *types.PreResultResp, err error) {
	// todo: add your logic here and delete this line

	return
}
