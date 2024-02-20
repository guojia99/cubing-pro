package public_player

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlayerResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlayerResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerResultLogic {
	return &PlayerResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlayerResultLogic) PlayerResult(req *types.PlayerResultReq) (resp *types.PlayerResultResp, err error) {
	// todo: add your logic here and delete this line

	return
}
