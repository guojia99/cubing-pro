package public_player

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlayerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlayerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerLogic {
	return &PlayerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlayerLogic) Player(req *types.PlayerReq) (resp *types.PlayerResp, err error) {
	// todo: add your logic here and delete this line

	return
}
