package public_player

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlayerSorLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlayerSorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerSorLogic {
	return &PlayerSorLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlayerSorLogic) PlayerSor(req *types.PlayerSorReq) (resp *types.PlayerSorResp, err error) {
	// todo: add your logic here and delete this line

	return
}
